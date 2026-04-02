package command

import (
	"testing"

	"github.com/jeffory/idotmatrix-lib/font"
	"github.com/jeffory/idotmatrix-lib/protocol"
)

func TestNewTextCommand(t *testing.T) {
	cmd := NewText("Hi",
		WithTextMode(protocol.TextScrollLeft),
		WithSpeed(128),
		WithColorMode(protocol.ColorFixed),
		WithTextColor(protocol.Color{R: 255, G: 0, B: 0}),
		WithFont(font.DefaultBitmap()),
		WithDisplaySize(protocol.Display32x32),
	)
	if !cmd.Chunked() { t.Error("Text command should be chunked") }
	packets, err := cmd.Encode()
	if err != nil { t.Fatalf("Encode() error: %v", err) }
	if len(packets) == 0 { t.Fatal("Encode() returned no packets") }

	p := packets[0]
	if p[2] != 0x03 { t.Errorf("opcode = 0x%02x, want 0x03", p[2]) }
	if p[15] != 0x0C { t.Errorf("fixed trailer = 0x%02x, want 0x0C", p[15]) }
}

func TestTextWithBackground(t *testing.T) {
	bg := protocol.Color{R: 0, G: 0, B: 255}
	cmd := NewText("A",
		WithTextColor(protocol.Color{R: 255, G: 0, B: 0}),
		WithBackground(bg),
		WithFont(font.DefaultBitmap()),
		WithDisplaySize(protocol.Display32x32),
	)
	packets, err := cmd.Encode()
	if err != nil { t.Fatalf("Encode() error: %v", err) }

	// Metadata bg mode is at offset 16 (header) + 10 (metadata index) = 26
	p := packets[0]
	if p[26] != 0x01 { t.Errorf("bg mode = 0x%02x, want 0x01", p[26]) }
}
