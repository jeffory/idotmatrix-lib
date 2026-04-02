package pixellab

import (
	"bytes"
	"image"
	"image/color"
	"image/gif"
	"testing"
)

func TestFramesToGIF(t *testing.T) {
	colors := []color.RGBA{
		{R: 255, A: 255},
		{G: 255, A: 255},
		{B: 255, A: 255},
		{R: 255, G: 255, A: 255},
	}

	frames := make([]image.Image, len(colors))
	for i, c := range colors {
		img := image.NewRGBA(image.Rect(0, 0, 64, 64))
		for y := range 64 {
			for x := range 64 {
				img.Set(x, y, c)
			}
		}
		frames[i] = img
	}

	data, err := FramesToGIF(frames, 10, nil)
	if err != nil {
		t.Fatal(err)
	}

	g, err := gif.DecodeAll(bytes.NewReader(data))
	if err != nil {
		t.Fatal(err)
	}

	if len(g.Image) != 4 {
		t.Fatalf("expected 4 frames, got %d", len(g.Image))
	}

	for i, frame := range g.Image {
		bounds := frame.Bounds()
		if bounds.Dx() != 64 || bounds.Dy() != 64 {
			t.Errorf("frame %d: expected 64x64, got %dx%d", i, bounds.Dx(), bounds.Dy())
		}
	}

	for i, d := range g.Delay {
		if d != 10 {
			t.Errorf("frame %d: expected delay 10, got %d", i, d)
		}
	}
}

func TestFramesToGIF_Empty(t *testing.T) {
	_, err := FramesToGIF(nil, 10, nil)
	if err == nil {
		t.Fatal("expected error for empty frames")
	}
}
