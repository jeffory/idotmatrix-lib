package font

import (
	"fmt"
	"image"
	"image/draw"
	"os"

	"github.com/jeffory/idotmatrix-lib/protocol"
	"golang.org/x/image/font"
	"golang.org/x/image/font/opentype"
	"golang.org/x/image/math/fixed"
)

type ttfRenderer struct {
	face font.Face
}

func FromTTF(fontData []byte, pointSize float64) (Renderer, error) {
	f, err := opentype.Parse(fontData)
	if err != nil {
		return nil, fmt.Errorf("parse font: %w", err)
	}

	face, err := opentype.NewFace(f, &opentype.FaceOptions{
		Size:    pointSize,
		DPI:     72,
		Hinting: font.HintingFull,
	})
	if err != nil {
		return nil, fmt.Errorf("create face: %w", err)
	}

	return &ttfRenderer{face: face}, nil
}

func FromTTFFile(path string, pointSize float64) (Renderer, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read font file: %w", err)
	}
	return FromTTF(data, pointSize)
}

func (r *ttfRenderer) RenderString(s string, size protocol.DisplaySize) ([][]byte, error) {
	var result [][]byte
	for _, ch := range s {
		bitmap := r.renderChar(ch, size)
		result = append(result, bitmap)
	}
	return result, nil
}

func (r *ttfRenderer) renderChar(ch rune, size protocol.DisplaySize) []byte {
	const (
		charWidth  = 16
		charHeight = 32
	)

	img := image.NewGray(image.Rect(0, 0, charWidth, charHeight))
	draw.Draw(img, img.Bounds(), image.Black, image.Point{}, draw.Src)

	advance, ok := r.face.GlyphAdvance(ch)
	if !ok {
		advance = fixed.I(charWidth)
	}

	textWidth := advance.Round()
	metrics := r.face.Metrics()
	ascent := metrics.Ascent.Round()

	textX := (charWidth - textWidth) / 2
	textY := (charHeight-metrics.Height.Round())/2 + ascent

	d := &font.Drawer{
		Dst:  img,
		Src:  image.White,
		Face: r.face,
		Dot:  fixed.P(textX, textY),
	}
	d.DrawString(string(ch))

	// Convert to row-major little-endian bitmap
	bitmapBytes := make([]byte, 64)
	bitPos := 0
	for y := 0; y < charHeight; y++ {
		for x := 0; x < charWidth; x++ {
			if img.GrayAt(x, y).Y > 128 {
				byteIdx := bitPos / 8
				bitIdx := bitPos % 8
				bitmapBytes[byteIdx] |= 1 << uint(bitIdx)
			}
			bitPos++
		}
	}

	result := make([]byte, 0, 68)
	result = append(result, separator...)
	result = append(result, bitmapBytes...)
	return result
}
