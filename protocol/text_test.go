package protocol

import (
	"encoding/binary"
	"testing"
)

func TestEncodeTextSingleChunk(t *testing.T) {
	// Single fake bitmap: 4-byte separator + 64 zero bytes = 68 bytes
	bitmap := make([]byte, 68)
	bitmap[0] = 0x05
	bitmap[1] = 0xFF
	bitmap[2] = 0xFF
	bitmap[3] = 0xFF

	opts := TextOptions{
		Mode:      TextFixed,
		Speed:     100,
		ColorMode: ColorFixed,
		TextColor: Color{R: 0xFF, G: 0x00, B: 0x00},
	}

	chunks, err := EncodeText([][]byte{bitmap}, opts)
	if err != nil {
		t.Fatalf("EncodeText() returned error: %v", err)
	}

	// Returns 1 chunk
	if len(chunks) != 1 {
		t.Fatalf("EncodeText() returned %d chunks, want 1", len(chunks))
	}

	pkt := chunks[0]

	// Opcode byte (index 2) = 0x03
	if pkt[2] != 0x03 {
		t.Errorf("opcode pkt[2] = 0x%02X, want 0x03", pkt[2])
	}

	// Fixed trailer (index 15) = 0x0C
	if pkt[15] != 0x0C {
		t.Errorf("trailer pkt[15] = 0x%02X, want 0x0C", pkt[15])
	}

	// Char count at metadata position: header is 16 bytes, metadata[0:2] = char count LE
	// pkt[16] = metadata[0], pkt[17] = metadata[1]
	if pkt[16] != 0x01 || pkt[17] != 0x00 {
		t.Errorf("char count pkt[16:18] = [0x%02X, 0x%02X], want [0x01, 0x00]", pkt[16], pkt[17])
	}

	// Text mode at metadata offset 4: pkt[16+4] = pkt[20] = 0x00 (TextFixed)
	if pkt[20] != 0x00 {
		t.Errorf("text mode pkt[20] = 0x%02X, want 0x00", pkt[20])
	}

	// Speed at metadata offset 5: pkt[16+5] = pkt[21] = 0x64 (100)
	if pkt[21] != 0x64 {
		t.Errorf("speed pkt[21] = 0x%02X, want 0x64", pkt[21])
	}

	// CRC in header (bytes 9-12) matches computed CRC32LE of metadata+bitmaps (bytes 16 onward)
	payload := pkt[16:]
	expectedCRC := CRC32LE(payload)
	if pkt[9] != expectedCRC[0] || pkt[10] != expectedCRC[1] || pkt[11] != expectedCRC[2] || pkt[12] != expectedCRC[3] {
		t.Errorf("CRC pkt[9:13] = %x, want %x", pkt[9:13], expectedCRC[:])
	}

	// Total length field (bytes 0-1, LE uint16) matches actual packet length
	totalLen := binary.LittleEndian.Uint16(pkt[0:2])
	if int(totalLen) != len(pkt) {
		t.Errorf("length field = %d, actual packet length = %d", totalLen, len(pkt))
	}
}

// TestEncodeTextReferencePacket uses the known-good reference packet structure from
// decoding_bytes.md. The reference header shows:
//
//	6200 03 00 00 52000000 <crc4b> 0000 0c
//
// where total length = 0x0062 = 98, payload length = 0x52 = 82 = 14 (metadata) + 68 (1 bitmap).
// We generate a packet with the reference parameters (1 char, mode=0, speed=100, colormode=1,
// color blue [00 00 ff], no background) and verify all structural fields are consistent.
func TestEncodeTextReferencePacket(t *testing.T) {
	// Reference metadata from decoding_bytes.md line 22:
	//   0100 00 01 00 64 01 00 00 ff 00 00 00 00
	// count=1, fixed=0x00 0x01, mode=0, speed=100(0x64), colormode=1,
	// R=0x00, G=0x00, B=0xFF, bgmode=0, bgR=0, bgG=0, bgB=0
	bitmap := make([]byte, 68)
	bitmap[0] = 0x05
	bitmap[1] = 0xFF
	bitmap[2] = 0xFF
	bitmap[3] = 0xFF

	opts := TextOptions{
		Mode:      TextFixed,
		Speed:     100,
		ColorMode: ColorFixed,
		TextColor: Color{R: 0x00, G: 0x00, B: 0xFF},
	}

	chunks, err := EncodeText([][]byte{bitmap}, opts)
	if err != nil {
		t.Fatalf("EncodeText() returned error: %v", err)
	}
	if len(chunks) != 1 {
		t.Fatalf("EncodeText() returned %d chunks, want 1", len(chunks))
	}

	pkt := chunks[0]

	// Reference: total length = 0x62 = 98 bytes (1 char * 68 bytes + 14 metadata + 16 header)
	if len(pkt) != 98 {
		t.Errorf("packet length = %d, want 98", len(pkt))
	}

	// CRC32 of bytes[16:] matches bytes[9:13]
	payloadBytes := pkt[16:]
	computedCRC := CRC32LE(payloadBytes)
	if pkt[9] != computedCRC[0] || pkt[10] != computedCRC[1] || pkt[11] != computedCRC[2] || pkt[12] != computedCRC[3] {
		t.Errorf("CRC mismatch: header[9:13] = %x, CRC32LE(pkt[16:]) = %x",
			pkt[9:13], computedCRC[:])
	}

	// Length field matches actual length
	lengthField := binary.LittleEndian.Uint16(pkt[0:2])
	if int(lengthField) != len(pkt) {
		t.Errorf("length field = %d, actual length = %d", lengthField, len(pkt))
	}

	// Payload size field (bytes 5-8) matches len(bytes[16:])
	payloadSizeField := binary.LittleEndian.Uint32(pkt[5:9])
	if int(payloadSizeField) != len(pkt[16:]) {
		t.Errorf("payload size field = %d, actual payload length = %d", payloadSizeField, len(pkt[16:]))
	}

	// Opcode = 0x03
	if pkt[2] != 0x03 {
		t.Errorf("opcode pkt[2] = 0x%02X, want 0x03", pkt[2])
	}

	// Trailer = 0x0C (matches reference `0c` at position 15)
	if pkt[15] != 0x0C {
		t.Errorf("trailer pkt[15] = 0x%02X, want 0x0C", pkt[15])
	}
}
