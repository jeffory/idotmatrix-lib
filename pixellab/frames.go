package pixellab

import (
	"bytes"
	"fmt"
	"image"
	"image/color/palette"
	"image/draw"
	"image/gif"

	xdraw "golang.org/x/image/draw"
)

// FramesToGIF encodes a slice of images as an animated GIF.
// delay is the inter-frame delay in 100ths of a second (e.g. 10 = 100ms).
// If size is non-nil, frames are scaled to the given dimensions.
func FramesToGIF(frames []image.Image, delay int, size *ImageSize) ([]byte, error) {
	if len(frames) == 0 {
		return nil, fmt.Errorf("no frames to encode")
	}

	plan9 := palette.Plan9
	palettedFrames := make([]*image.Paletted, len(frames))
	delays := make([]int, len(frames))

	for i, frame := range frames {
		src := frame
		if size != nil {
			scaled := image.NewRGBA(image.Rect(0, 0, size.Width, size.Height))
			xdraw.NearestNeighbor.Scale(scaled, scaled.Bounds(), frame, frame.Bounds(), xdraw.Over, nil)
			src = scaled
		}

		bounds := src.Bounds()
		p := image.NewPaletted(image.Rect(0, 0, bounds.Dx(), bounds.Dy()), plan9)
		draw.FloydSteinberg.Draw(p, p.Bounds(), src, bounds.Min)
		palettedFrames[i] = p
		delays[i] = delay
	}

	var buf bytes.Buffer
	err := gif.EncodeAll(&buf, &gif.GIF{
		Image:     palettedFrames,
		Delay:     delays,
		LoopCount: 0, // loop forever
	})
	if err != nil {
		return nil, fmt.Errorf("encode GIF: %w", err)
	}

	return buf.Bytes(), nil
}
