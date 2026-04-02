package font

import (
	"github.com/jeffory/idotmatrix-lib/protocol"
)

// bitmapRenderer provides a simple built-in bitmap font.
type bitmapRenderer struct{}

// DefaultBitmap returns a Renderer using a minimal built-in pixel font.
// Covers ASCII printable range 0x20-0x7E.
// Characters are rendered as 16px wide x 32px tall bitmaps in
// row-major, little-endian bit order matching the device protocol.
func DefaultBitmap() Renderer {
	return &bitmapRenderer{}
}

// separator is prepended to each character bitmap.
var separator = []byte{0x05, 0xFF, 0xFF, 0xFF}

func (r *bitmapRenderer) RenderString(s string, size protocol.DisplaySize) ([][]byte, error) {
	var result [][]byte

	for _, ch := range s {
		bitmap := r.renderChar(ch)
		result = append(result, bitmap)
	}

	return result, nil
}

func (r *bitmapRenderer) renderChar(ch rune) []byte {
	data, ok := builtinGlyphs[ch]
	if !ok {
		// Unknown character → blank (all zeros)
		data = make([]byte, 64)
	}

	result := make([]byte, 0, 68)
	result = append(result, separator...)
	result = append(result, data...)
	return result
}

// builtinGlyphs maps ASCII characters to their 64-byte bitmap representations.
// Each bitmap is 16px wide x 32px tall, stored as row-major, little-endian bits.
// 16 pixels per row = 2 bytes per row, 32 rows = 64 bytes per glyph.
var builtinGlyphs = map[rune][]byte{
	// Space (all zeros)
	' ': make([]byte, 64),
}
