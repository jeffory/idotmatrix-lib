package command

import (
	"encoding/hex"
	"testing"
	"time"

	"github.com/jeffory/idotmatrix-lib/protocol"
)

func TestPowerOnCommand(t *testing.T) {
	cmd := PowerOn()
	if cmd.Chunked() { t.Error("PowerOn should not be chunked") }
	packets, err := cmd.Encode()
	if err != nil { t.Fatalf("Encode() error: %v", err) }
	if len(packets) != 1 { t.Fatalf("got %d packets, want 1", len(packets)) }
	want, _ := hex.DecodeString("0500070101")
	if !bytesEq(packets[0], want) { t.Errorf("packet = %x, want %x", packets[0], want) }
}

func TestPowerOffCommand(t *testing.T) {
	cmd := PowerOff()
	packets, _ := cmd.Encode()
	want, _ := hex.DecodeString("0500070100")
	if !bytesEq(packets[0], want) { t.Errorf("packet = %x, want %x", packets[0], want) }
}

func TestResetCommand(t *testing.T) {
	cmd := Reset()
	if cmd.Chunked() { t.Error("Reset should not be chunked") }
	packets, err := cmd.Encode()
	if err != nil { t.Fatalf("Encode() error: %v", err) }
	if len(packets) != 2 { t.Fatalf("got %d packets, want 2", len(packets)) }
	wantP1, _ := hex.DecodeString("04000380")
	wantP2, _ := hex.DecodeString("0500048050")
	if !bytesEq(packets[0], wantP1) { t.Errorf("packet1 = %x, want %x", packets[0], wantP1) }
	if !bytesEq(packets[1], wantP2) { t.Errorf("packet2 = %x, want %x", packets[1], wantP2) }
}

func TestSyncTimeCommand(t *testing.T) {
	loc := time.FixedZone("test", 0)
	ts := time.Date(2023, 12, 18, 10, 38, 16, 0, loc)
	cmd := SyncTime(ts)
	packets, err := cmd.Encode()
	if err != nil { t.Fatalf("Encode() error: %v", err) }
	want, _ := hex.DecodeString("0b000180e70c12010a2610")
	if !bytesEq(packets[0], want) { t.Errorf("packet = %x, want %x", packets[0], want) }
}

func TestDrawPixelCommand(t *testing.T) {
	cmd := DrawPixel(31, 31, protocol.Color{R: 255, G: 0, B: 0})
	if cmd.Chunked() { t.Error("DrawPixel should not be chunked") }
	packets, err := cmd.Encode()
	if err != nil { t.Fatalf("Encode() error: %v", err) }
	want, _ := hex.DecodeString("0a00050100ff00001f1f")
	if !bytesEq(packets[0], want) { t.Errorf("packet = %x, want %x", packets[0], want) }
}

func bytesEq(a, b []byte) bool {
	if len(a) != len(b) { return false }
	for i := range a {
		if a[i] != b[i] { return false }
	}
	return true
}
