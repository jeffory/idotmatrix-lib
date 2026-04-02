package command

import (
	"bytes"
	"fmt"
	"image"
	"image/color/palette"
	"image/gif"
	_ "image/jpeg"
	_ "image/png"
	"os"
	"path/filepath"
	"strings"

	"image/draw"

	"github.com/jeffory/idotmatrix-lib/protocol"
	xdraw "golang.org/x/image/draw"
)

// ImageOptions configures image processing before display.
type ImageOptions struct {
	Contrast   float64
	Saturation float64
}

// ImageOption is a functional option for image processing.
type ImageOption func(*ImageOptions)

// WithContrast sets the contrast adjustment (-100 to 100).
func WithContrast(v float64) ImageOption {
	return func(o *ImageOptions) { o.Contrast = v }
}

// WithSaturation sets the saturation adjustment (-100 to 100).
func WithSaturation(v float64) ImageOption {
	return func(o *ImageOptions) { o.Saturation = v }
}

// NewImage creates a display command from a decoded image. The image is scaled
// to the display size, optionally adjusted, and encoded as a GIF.
func NewImage(src image.Image, size protocol.DisplaySize, opts ...ImageOption) (Command, error) {
	var options ImageOptions
	for _, o := range opts {
		o(&options)
	}

	dst := image.NewRGBA(image.Rect(0, 0, size.Width, size.Height))
	xdraw.BiLinear.Scale(dst, dst.Bounds(), src, src.Bounds(), xdraw.Over, nil)

	AdjustImage(dst, options.Contrast, options.Saturation)

	var buf bytes.Buffer
	err := gif.Encode(&buf, dst, &gif.Options{NumColors: 256})
	if err != nil {
		return nil, fmt.Errorf("encode GIF: %w", err)
	}

	return NewGIF(&buf, size)
}

// NewImageFromFile loads an image file (PNG, JPEG, or GIF) and creates a
// display command from it. Animated GIFs are preserved as animations with
// frames scaled to the display size.
func NewImageFromFile(path string, size protocol.DisplaySize, opts ...ImageOption) (Command, error) {
	// Check if this is an animated GIF.
	if isGIFFile(path) {
		cmd, err := tryAnimatedGIF(path, size, opts...)
		if err != nil {
			return nil, err
		}
		if cmd != nil {
			return cmd, nil
		}
		// Single-frame GIF falls through to static path.
	}

	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open image: %w", err)
	}
	defer f.Close()

	src, _, err := image.Decode(f)
	if err != nil {
		return nil, fmt.Errorf("decode image: %w", err)
	}

	return NewImage(src, size, opts...)
}

// isGIFFile checks if the file has a .gif extension.
func isGIFFile(path string) bool {
	return strings.EqualFold(filepath.Ext(path), ".gif")
}

// tryAnimatedGIF attempts to load an animated GIF. Returns nil command if the
// GIF has only one frame.
func tryAnimatedGIF(path string, size protocol.DisplaySize, opts ...ImageOption) (Command, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open GIF: %w", err)
	}
	defer f.Close()

	g, err := gif.DecodeAll(f)
	if err != nil {
		return nil, fmt.Errorf("decode GIF: %w", err)
	}

	if len(g.Image) <= 1 {
		return nil, nil
	}

	var options ImageOptions
	for _, o := range opts {
		o(&options)
	}

	return scaleAnimatedGIF(g, size, options)
}

// scaleAnimatedGIF scales all frames of an animated GIF to the target size,
// compositing partial frames onto a full canvas and handling disposal.
func scaleAnimatedGIF(g *gif.GIF, size protocol.DisplaySize, options ImageOptions) (Command, error) {
	srcW := g.Config.Width
	srcH := g.Config.Height
	if srcW == 0 || srcH == 0 {
		b := g.Image[0].Bounds()
		srcW = b.Dx()
		srcH = b.Dy()
	}

	canvas := image.NewRGBA(image.Rect(0, 0, srcW, srcH))
	plan9 := palette.Plan9

	scaledFrames := make([]*image.Paletted, len(g.Image))

	for i, frame := range g.Image {
		// Draw this frame onto the canvas at its offset.
		draw.Draw(canvas, frame.Bounds(), frame, frame.Bounds().Min, draw.Over)

		// Scale the full canvas to display size.
		scaled := image.NewRGBA(image.Rect(0, 0, size.Width, size.Height))
		xdraw.NearestNeighbor.Scale(scaled, scaled.Bounds(), canvas, canvas.Bounds(), xdraw.Over, nil)

		// Apply adjustments.
		AdjustImage(scaled, options.Contrast, options.Saturation)

		// Convert to paletted.
		p := image.NewPaletted(scaled.Bounds(), plan9)
		draw.FloydSteinberg.Draw(p, p.Bounds(), scaled, scaled.Bounds().Min)
		scaledFrames[i] = p

		// Handle disposal.
		if i < len(g.Disposal) {
			switch g.Disposal[i] {
			case gif.DisposalBackground:
				draw.Draw(canvas, frame.Bounds(), image.Transparent, image.Point{}, draw.Src)
			case gif.DisposalPrevious:
				// For simplicity, treat as background disposal.
				draw.Draw(canvas, frame.Bounds(), image.Transparent, image.Point{}, draw.Src)
			}
			// DisposalNone: leave canvas as-is.
		}
	}

	var buf bytes.Buffer
	err := gif.EncodeAll(&buf, &gif.GIF{
		Image:     scaledFrames,
		Delay:     g.Delay,
		LoopCount: g.LoopCount,
	})
	if err != nil {
		return nil, fmt.Errorf("encode animated GIF: %w", err)
	}

	return NewGIF(&buf, size)
}
