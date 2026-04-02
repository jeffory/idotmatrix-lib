package protocol

// EncodeGraffiti returns the 10-byte packet to draw a single pixel.
// Coordinates are 0-based. Values beyond the display size are clamped by the device.
func EncodeGraffiti(c Color, x, y uint8) []byte {
	return []byte{
		0x0A, 0x00, // length (10 bytes, LE)
		0x05, 0x01, 0x00, // fixed header
		c.R, c.G, c.B,
		x, y,
	}
}
