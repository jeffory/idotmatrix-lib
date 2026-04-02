package font

import "github.com/jeffory/idotmatrix-lib/protocol"

// separator is prepended to each character bitmap.
var separator = []byte{0x05, 0xFF, 0xFF, 0xFF}

// Renderer converts text to bitmaps suitable for the iDotMatrix protocol.
// Each bitmap is returned as a byte slice with a 4-byte separator prefix
// [0x05, 0xFF, 0xFF, 0xFF] followed by row-major, little-endian bitmap data.
type Renderer interface {
	RenderString(s string, size protocol.DisplaySize) ([][]byte, error)
}
