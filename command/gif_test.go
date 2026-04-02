package command

import (
	"bytes"
	"testing"

	"github.com/jeffory/idotmatrix-lib/protocol"
)

func TestNewGIF(t *testing.T) {
	gifData := make([]byte, 200)
	copy(gifData, []byte("GIF89a"))
	gifData[len(gifData)-1] = 0x3B

	cmd, err := NewGIF(bytes.NewReader(gifData), protocol.Display32x32)
	if err != nil { t.Fatalf("NewGIF() error: %v", err) }
	if !cmd.Chunked() { t.Error("GIF command should be chunked") }
	packets, err := cmd.Encode()
	if err != nil { t.Fatalf("Encode() error: %v", err) }
	if len(packets) != 1 { t.Errorf("got %d packets, want 1", len(packets)) }
}
