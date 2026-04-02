package font

import (
	"testing"

	"github.com/jeffory/idotmatrix-lib/protocol"
)

func TestDefaultBitmapRenderString(t *testing.T) {
	r := DefaultBitmap()

	bitmaps, err := r.RenderString("A", protocol.Display32x32)
	if err != nil {
		t.Fatalf("RenderString() error: %v", err)
	}

	if len(bitmaps) != 1 {
		t.Fatalf("got %d bitmaps, want 1", len(bitmaps))
	}

	// Each bitmap: 4-byte separator + 64-byte bitmap = 68 bytes
	if len(bitmaps[0]) != 68 {
		t.Errorf("bitmap length = %d, want 68", len(bitmaps[0]))
	}

	// Separator check
	if bitmaps[0][0] != 0x05 || bitmaps[0][1] != 0xFF ||
		bitmaps[0][2] != 0xFF || bitmaps[0][3] != 0xFF {
		t.Errorf("separator = %x, want 05ffffff", bitmaps[0][:4])
	}

	// "A" should have non-zero pixel data
	hasPixels := false
	for _, b := range bitmaps[0][4:] {
		if b != 0 {
			hasPixels = true
			break
		}
	}
	if !hasPixels {
		t.Error("bitmap for 'A' is all zeros, expected visible pixels")
	}
}

func TestDefaultBitmapMultipleChars(t *testing.T) {
	r := DefaultBitmap()

	bitmaps, err := r.RenderString("Hi", protocol.Display32x32)
	if err != nil {
		t.Fatalf("RenderString() error: %v", err)
	}

	if len(bitmaps) != 2 {
		t.Fatalf("got %d bitmaps, want 2", len(bitmaps))
	}

	for i, bm := range bitmaps {
		if len(bm) != 68 {
			t.Errorf("bitmap[%d] length = %d, want 68", i, len(bm))
		}
		// Each character should have non-zero pixel data
		hasPixels := false
		for _, b := range bm[4:] {
			if b != 0 {
				hasPixels = true
				break
			}
		}
		if !hasPixels {
			t.Errorf("bitmap[%d] is all zeros, expected visible pixels", i)
		}
	}
}

func TestDefaultBitmapSpaceIsBlank(t *testing.T) {
	r := DefaultBitmap()

	bitmaps, err := r.RenderString(" ", protocol.Display32x32)
	if err != nil {
		t.Fatalf("RenderString() error: %v", err)
	}

	// Space bitmap (after separator) should be all zeros
	bitmap := bitmaps[0][4:]
	for i, b := range bitmap {
		if b != 0 {
			t.Errorf("space bitmap byte %d = 0x%02x, want 0x00", i, b)
		}
	}
}
