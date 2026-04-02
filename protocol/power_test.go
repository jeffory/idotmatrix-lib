package protocol

import (
	"encoding/hex"
	"testing"
)

func TestEncodePowerOn(t *testing.T) {
	got := EncodePowerOn()
	want, _ := hex.DecodeString("0500070101")
	if !bytesEqual(got, want) {
		t.Errorf("EncodePowerOn() = %x, want %x", got, want)
	}
}

func TestEncodePowerOff(t *testing.T) {
	got := EncodePowerOff()
	want, _ := hex.DecodeString("0500070100")
	if !bytesEqual(got, want) {
		t.Errorf("EncodePowerOff() = %x, want %x", got, want)
	}
}

func TestEncodeReset(t *testing.T) {
	part1, part2 := EncodeReset()
	wantPart1, _ := hex.DecodeString("04000380")
	wantPart2, _ := hex.DecodeString("0500048050")

	if !bytesEqual(part1, wantPart1) {
		t.Errorf("EncodeReset() part1 = %x, want %x", part1, wantPart1)
	}
	if !bytesEqual(part2, wantPart2) {
		t.Errorf("EncodeReset() part2 = %x, want %x", part2, wantPart2)
	}
}

func bytesEqual(a, b []byte) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
