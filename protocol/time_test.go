package protocol

import (
	"encoding/hex"
	"testing"
	"time"
)

func TestEncodeTimeSync(t *testing.T) {
	// Reference from BT snoop: 0b000180e70c12010a2610
	// year=0xe7(231=2023&0xff), month=12, day=18, dow=1(Mon), hour=10, min=38, sec=16
	loc := time.FixedZone("test", 0)
	ts := time.Date(2023, 12, 18, 10, 38, 16, 0, loc) // Monday

	got := EncodeTimeSync(ts)
	want, _ := hex.DecodeString("0b000180e70c12010a2610")

	if len(got) != len(want) {
		t.Fatalf("length mismatch: got %d, want %d", len(got), len(want))
	}
	for i := range got {
		if got[i] != want[i] {
			t.Errorf("byte %d: got 0x%02x, want 0x%02x", i, got[i], want[i])
		}
	}
}

func TestEncodeTimeSyncLength(t *testing.T) {
	got := EncodeTimeSync(time.Now())
	if len(got) != 11 {
		t.Errorf("EncodeTimeSync() length = %d, want 11", len(got))
	}
}
