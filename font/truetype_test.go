package font

import (
	"os"
	"testing"

	"github.com/jeffory/idotmatrix-lib/protocol"
)

func TestFromTTFFile(t *testing.T) {
	fontPaths := []string{
		"/usr/share/fonts/truetype/dejavu/DejaVuSans-Bold.ttf",
		"/usr/share/fonts/dejavu-sans-fonts/DejaVuSans-Bold.ttf",
		"/usr/share/fonts/TTF/DejaVuSans-Bold.ttf",
		"/usr/share/fonts/truetype/dejavu/DejaVuSans.ttf",
		"/usr/share/fonts/dejavu-sans-fonts/DejaVuSans.ttf",
		// Fedora / GNOME Adwaita fonts
		"/usr/share/fonts/adwaita-sans-fonts/AdwaitaSans-Regular.ttf",
		"/usr/share/fonts/adwaita-mono-fonts/AdwaitaMono-Regular.ttf",
		// Fedora Carlito (metric-compatible with Calibri)
		"/usr/share/fonts/google-carlito-fonts/Carlito-Regular.ttf",
	}

	var fontPath string
	for _, p := range fontPaths {
		if _, err := os.Stat(p); err == nil {
			fontPath = p
			break
		}
	}

	if fontPath == "" {
		t.Skip("no system TrueType font found")
	}

	r, err := FromTTFFile(fontPath, 24)
	if err != nil {
		t.Fatalf("FromTTFFile() error: %v", err)
	}

	bitmaps, err := r.RenderString("AB", protocol.Display32x32)
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
		if bm[0] != 0x05 || bm[1] != 0xFF || bm[2] != 0xFF || bm[3] != 0xFF {
			t.Errorf("bitmap[%d] separator = %x, want 05ffffff", i, bm[:4])
		}
	}

	// A and B should produce different bitmaps
	aBits := bitmaps[0][4:]
	bBits := bitmaps[1][4:]
	same := true
	for i := range aBits {
		if aBits[i] != bBits[i] {
			same = false
			break
		}
	}
	if same {
		t.Error("A and B produced identical bitmaps")
	}

	// Bitmap should not be all zeros
	allZero := true
	for _, b := range bitmaps[0][4:] {
		if b != 0 {
			allZero = false
			break
		}
	}
	if allZero {
		t.Error("bitmap for 'A' is all zeros")
	}
}
