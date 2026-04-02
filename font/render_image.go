package font

import (
	"image"
	"image/color"
	"image/draw"

	"github.com/jeffory/idotmatrix-lib/protocol"
	xdraw "golang.org/x/image/font"
	"golang.org/x/image/math/fixed"
)

// RenderStringImage renders the full text string into a single RGBA image
// with the given foreground color and display height. The width is sized to
// fit all characters. Only works with ttfRenderer.
func RenderStringImage(r Renderer, s string, size protocol.DisplaySize, fg color.RGBA) *image.RGBA {
	ttf, ok := r.(*ttfRenderer)
	if !ok {
		// Fallback: render character-by-character into a basic image
		return renderFallbackImage(r, s, size, fg)
	}

	height := size.Height

	// Measure total text width
	totalWidth := 0
	for _, ch := range s {
		adv, ok := ttf.face.GlyphAdvance(ch)
		if !ok {
			totalWidth += 16
		} else {
			totalWidth += adv.Round()
		}
	}
	if totalWidth == 0 {
		totalWidth = 1
	}

	img := image.NewRGBA(image.Rect(0, 0, totalWidth, height))
	draw.Draw(img, img.Bounds(), image.Black, image.Point{}, draw.Src)

	metrics := ttf.face.Metrics()
	ascent := metrics.Ascent.Round()
	textY := (height-metrics.Height.Round())/2 + ascent

	d := &xdraw.Drawer{
		Dst:  img,
		Src:  image.NewUniform(fg),
		Face: ttf.face,
		Dot:  fixed.P(0, textY),
	}
	d.DrawString(s)

	return img
}

func renderFallbackImage(r Renderer, s string, size protocol.DisplaySize, fg color.RGBA) *image.RGBA {
	height := size.Height
	charWidth := 16
	width := len([]rune(s)) * charWidth

	img := image.NewRGBA(image.Rect(0, 0, width, height))
	draw.Draw(img, img.Bounds(), image.Black, image.Point{}, draw.Src)

	return img
}
