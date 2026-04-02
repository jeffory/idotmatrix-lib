package command

import (
	"fmt"
	"io"
	"os"

	"github.com/jeffory/idotmatrix-lib/protocol"
)

type gifCmd struct {
	data []byte
	size protocol.DisplaySize
}

func NewGIF(r io.Reader, size protocol.DisplaySize) (Command, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("read GIF data: %w", err)
	}
	return &gifCmd{data: data, size: size}, nil
}

func NewGIFFromFile(path string, size protocol.DisplaySize) (Command, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open GIF file: %w", err)
	}
	defer f.Close()
	return NewGIF(f, size)
}

func (*gifCmd) Chunked() bool { return true }

func (c *gifCmd) Encode() ([][]byte, error) {
	return protocol.EncodeGIF(c.data, c.size)
}
