package command

import (
	"bytes"
	"fmt"
	"image"
	"image/gif"
	_ "image/jpeg"
	_ "image/png"
	"os"

	"github.com/jeffory/idotmatrix-lib/protocol"
	"golang.org/x/image/draw"
)

func NewImageFromFile(path string, size protocol.DisplaySize) (Command, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open image: %w", err)
	}
	defer f.Close()

	src, _, err := image.Decode(f)
	if err != nil {
		return nil, fmt.Errorf("decode image: %w", err)
	}

	dst := image.NewRGBA(image.Rect(0, 0, size.Width, size.Height))
	draw.BiLinear.Scale(dst, dst.Bounds(), src, src.Bounds(), draw.Over, nil)

	palettedImg := image.NewPaletted(dst.Bounds(), nil)
	draw.FloydSteinberg.Draw(palettedImg, dst.Bounds(), dst, image.Point{})

	var buf bytes.Buffer
	err = gif.Encode(&buf, palettedImg, nil)
	if err != nil {
		return nil, fmt.Errorf("encode GIF: %w", err)
	}

	return NewGIF(&buf, size)
}
