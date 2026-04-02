package protocol

import (
	"encoding/hex"
	"testing"
)

func TestEncodeGraffiti(t *testing.T) {
	tests := []struct {
		name    string
		color   Color
		x, y    uint8
		wantHex string
	}{
		{
			name:    "red pixel at origin",
			color:   Color{R: 255, G: 0, B: 0},
			x:       0, y: 0,
			wantHex: "0a00050100ff00000000",
		},
		{
			name:    "red pixel at max corner",
			color:   Color{R: 255, G: 0, B: 0},
			x:       31, y: 31,
			wantHex: "0a00050100ff00001f1f",
		},
		{
			name:    "red pixel at (30,31) from BT snoop",
			color:   Color{R: 255, G: 0, B: 0},
			x:       30, y: 31,
			wantHex: "0a00050100ff00001e1f",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := EncodeGraffiti(tt.color, tt.x, tt.y)
			want, _ := hex.DecodeString(tt.wantHex)
			if len(got) != len(want) {
				t.Fatalf("length: got %d, want %d", len(got), len(want))
			}
			for i := range got {
				if got[i] != want[i] {
					t.Errorf("byte %d: got 0x%02x, want 0x%02x", i, got[i], want[i])
				}
			}
		})
	}
}
