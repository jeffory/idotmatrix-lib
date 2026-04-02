package command

import "github.com/jeffory/idotmatrix-lib/protocol"

type drawPixelCmd struct {
	x, y  uint8
	color protocol.Color
}

func DrawPixel(x, y uint8, c protocol.Color) Command { return &drawPixelCmd{x: x, y: y, color: c} }
func (*drawPixelCmd) Chunked() bool                   { return false }
func (c *drawPixelCmd) Encode() ([][]byte, error) {
	return [][]byte{protocol.EncodeGraffiti(c.color, c.x, c.y)}, nil
}
