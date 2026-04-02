package font

import (
	"log"
)

// DefaultBitmap returns a Renderer using the embedded Go Mono font,
// rasterized at a size suitable for the 16x32 pixel character grid.
func DefaultBitmap() Renderer {
	r, err := FromTTF(embeddedFontData, 24)
	if err != nil {
		log.Fatalf("font: failed to load embedded font: %v", err)
	}
	return r
}
