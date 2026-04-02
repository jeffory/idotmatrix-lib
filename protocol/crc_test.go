package protocol

import (
	"encoding/hex"
	"testing"
)

func TestCRC32LE(t *testing.T) {
	tests := []struct {
		name     string
		inputHex string
		wantHex  string
	}{
		{
			name:     "text metadata+bitmaps from reference capture",
			inputHex: "070000010000010101010000000002ffffff000000000302021a26424242261a000002ffffff000000000000003c42424242423c000002ffffff000000000008083e080808084830000002ffffff000000000008083e080808084830000002ffffff000000000000003c42424242423c000002ffffff000000000000007f9292929292b7000002ffffff000000000000007c42023c40423e0000",
			wantHex:  "fad1c25c",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			input, err := hex.DecodeString(tt.inputHex)
			if err != nil {
				t.Fatalf("bad input hex: %v", err)
			}
			want, err := hex.DecodeString(tt.wantHex)
			if err != nil {
				t.Fatalf("bad want hex: %v", err)
			}

			got := CRC32LE(input)
			if got[0] != want[0] || got[1] != want[1] || got[2] != want[2] || got[3] != want[3] {
				t.Errorf("CRC32LE() = %x, want %x", got, want)
			}
		})
	}
}
