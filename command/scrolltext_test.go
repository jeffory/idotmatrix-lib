package command

import (
	"testing"

	"github.com/jeffory/idotmatrix-lib/protocol"
)

func TestScrollTextEncode(t *testing.T) {
	c := NewScrollText("Hello",
		WithScrollTextColor(protocol.Color{R: 255, G: 255, B: 255}),
		WithScrollDisplaySize(protocol.Display32x32),
	)

	chunks, err := c.Encode()
	if err != nil {
		t.Fatalf("Encode() error: %v", err)
	}

	if len(chunks) == 0 {
		t.Fatal("expected at least one chunk")
	}

	// Each chunk should have a 16-byte GIF header
	for i, chunk := range chunks {
		if len(chunk) < 16 {
			t.Errorf("chunk[%d] too short: %d bytes", i, len(chunk))
		}
		// GIF opcode byte 15 should be 0x0D
		if chunk[15] != 0x0D {
			t.Errorf("chunk[%d] byte 15 = 0x%02x, want 0x0D", i, chunk[15])
		}
	}
}

func TestScrollTextChunked(t *testing.T) {
	c := NewScrollText("Hi")
	if !c.Chunked() {
		t.Error("Chunked() = false, want true")
	}
}
