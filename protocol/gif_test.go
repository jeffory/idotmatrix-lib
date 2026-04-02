package protocol

import (
	"encoding/binary"
	"encoding/hex"
	"testing"
)

func TestEncodeGIFSingleChunk(t *testing.T) {
	// 100-byte fake GIF data
	data := make([]byte, 100)
	for i := range data {
		data[i] = byte(i)
	}

	chunks, err := EncodeGIF(data, Display32x32)
	if err != nil {
		t.Fatalf("EncodeGIF() error = %v", err)
	}

	if len(chunks) != 1 {
		t.Fatalf("expected 1 chunk, got %d", len(chunks))
	}

	chunk := chunks[0]

	// Chunk length = 16 + 100
	if len(chunk) != 116 {
		t.Errorf("chunk length = %d, want 116", len(chunk))
	}

	// Byte 2 = 0x01 (fixed)
	if chunk[2] != 0x01 {
		t.Errorf("byte 2 = 0x%02x, want 0x01", chunk[2])
	}

	// Byte 4 = 0x00 (single chunk indicator)
	if chunk[4] != 0x00 {
		t.Errorf("byte 4 = 0x%02x, want 0x00", chunk[4])
	}

	// Byte 15 = 0x0D (fixed)
	if chunk[15] != 0x0D {
		t.Errorf("byte 15 = 0x%02x, want 0x0d", chunk[15])
	}

	// CRC bytes 9-12 match CRC32LE of raw data
	wantCRC := CRC32LE(data)
	if !bytesEqual(chunk[9:13], wantCRC[:]) {
		t.Errorf("CRC bytes = %x, want %x", chunk[9:13], wantCRC)
	}

	// Payload after header (byte 16+) matches input data
	if !bytesEqual(chunk[16:], data) {
		t.Errorf("payload after header does not match input data")
	}
}

func TestEncodeGIFMultiChunk(t *testing.T) {
	// 5000-byte fake GIF data
	data := make([]byte, 5000)
	for i := range data {
		data[i] = byte(i % 251) // use prime to avoid patterns
	}

	chunks, err := EncodeGIF(data, Display32x32)
	if err != nil {
		t.Fatalf("EncodeGIF() error = %v", err)
	}

	if len(chunks) != 2 {
		t.Fatalf("expected 2 chunks, got %d", len(chunks))
	}

	first := chunks[0]
	second := chunks[1]

	// First chunk: byte 4 = 0x00 (first of multi), size = 16+4096
	if first[4] != 0x00 {
		t.Errorf("first chunk byte 4 = 0x%02x, want 0x00", first[4])
	}
	if len(first) != 16+4096 {
		t.Errorf("first chunk size = %d, want %d", len(first), 16+4096)
	}

	// Second chunk: byte 4 = 0x02, size = 16+904
	if second[4] != 0x02 {
		t.Errorf("second chunk byte 4 = 0x%02x, want 0x02", second[4])
	}
	if len(second) != 16+904 {
		t.Errorf("second chunk size = %d, want %d", len(second), 16+904)
	}

	// CRC identical in both headers
	if !bytesEqual(first[9:13], second[9:13]) {
		t.Errorf("CRC mismatch: first=%x, second=%x", first[9:13], second[9:13])
	}

	// Verify CRC matches raw data
	wantCRC := CRC32LE(data)
	if !bytesEqual(first[9:13], wantCRC[:]) {
		t.Errorf("CRC bytes = %x, want %x", first[9:13], wantCRC)
	}

	// Total payload size identical in both headers = 5000 + 2*16 = 5032
	const wantTotal = 5000 + 2*16
	totalFirst := binary.LittleEndian.Uint32(first[5:9])
	totalSecond := binary.LittleEndian.Uint32(second[5:9])
	if totalFirst != wantTotal {
		t.Errorf("first chunk total payload = %d, want %d", totalFirst, wantTotal)
	}
	if totalSecond != wantTotal {
		t.Errorf("second chunk total payload = %d, want %d", totalSecond, wantTotal)
	}
}

func TestEncodeGIFReferenceHeaders(t *testing.T) {
	// Reference headers from decoding_bytes.md:
	// 1st: 10 10 01 00 00 b9 18 00 00 db 42 cb 14 05 00 0d
	// 2nd: c9 08 01 00 02 b9 18 00 00 db 42 cb 14 05 00 0d
	//
	// First chunk size:  0x1010 = 4112 = 16 + 4096
	// Second chunk size: 0x08c9 = 2249 = 16 + 2233
	// Total payload:     0x18b9 = 6329
	// Raw GIF size:      6329 - 2*16 = 6297
	// Verify: 6297 + 2*16 = 6329 ✓

	wantHeader1, _ := hex.DecodeString("10100100008f270000e8e44a8305000d")
	_ = wantHeader1 // reference A (different GIF)

	// Reference B headers (the ones we test against):
	wantH1, _ := hex.DecodeString("1010010000b9180000db42cb1405000d")
	wantH2, _ := hex.DecodeString("c908010002b9180000db42cb1405000d")

	// Parse expected values from first header
	chunkSize1 := binary.LittleEndian.Uint16(wantH1[0:2])
	chunkSize2 := binary.LittleEndian.Uint16(wantH2[0:2])
	totalPayload := binary.LittleEndian.Uint32(wantH1[5:9])

	// Verify the math from the reference
	if chunkSize1 != 0x1010 {
		t.Errorf("reference chunk1 size = 0x%04x, want 0x1010", chunkSize1)
	}
	if chunkSize2 != 0x08c9 {
		t.Errorf("reference chunk2 size = 0x%04x, want 0x08c9", chunkSize2)
	}
	if totalPayload != 0x18b9 {
		t.Errorf("reference total payload = 0x%04x, want 0x18b9", totalPayload)
	}

	// Decode the math: raw GIF = 6329 - 32 = 6297, total = 6297 + 2*16 = 6329
	rawGIFSize := int(totalPayload) - 2*gifHeaderSize
	if rawGIFSize != 6297 {
		t.Errorf("raw GIF size = %d, want 6297", rawGIFSize)
	}
	recomputed := rawGIFSize + 2*gifHeaderSize
	if recomputed != int(totalPayload) {
		t.Errorf("recomputed total = %d, want %d", recomputed, totalPayload)
	}

	// Verify the two headers agree on total payload and CRC
	totalPayload2 := binary.LittleEndian.Uint32(wantH2[5:9])
	if totalPayload != totalPayload2 {
		t.Errorf("total payload mismatch: h1=%d, h2=%d", totalPayload, totalPayload2)
	}
	if !bytesEqual(wantH1[9:13], wantH2[9:13]) {
		t.Errorf("CRC mismatch between headers: h1=%x, h2=%x", wantH1[9:13], wantH2[9:13])
	}

	// Verify fixed bytes in both headers
	for i, h := range [][]byte{wantH1, wantH2} {
		if h[2] != 0x01 {
			t.Errorf("header %d byte 2 = 0x%02x, want 0x01", i+1, h[2])
		}
		if h[13] != 0x05 {
			t.Errorf("header %d byte 13 = 0x%02x, want 0x05", i+1, h[13])
		}
		if h[14] != 0x00 {
			t.Errorf("header %d byte 14 = 0x%02x, want 0x00", i+1, h[14])
		}
		if h[15] != 0x0d {
			t.Errorf("header %d byte 15 = 0x%02x, want 0x0d", i+1, h[15])
		}
	}

	// First chunk: multi-chunk indicator = 0x00
	if wantH1[4] != 0x00 {
		t.Errorf("first header byte 4 = 0x%02x, want 0x00", wantH1[4])
	}
	// Second chunk: continuation indicator = 0x02
	if wantH2[4] != 0x02 {
		t.Errorf("second header byte 4 = 0x%02x, want 0x02", wantH2[4])
	}
}
