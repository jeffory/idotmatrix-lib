package protocol

import (
	"encoding/binary"
	"fmt"
)

const (
	textHeaderSize   = 16
	textMetadataSize = 14
)

// EncodeText builds the complete text packet(s) from pre-rendered character bitmaps.
// Each bitmap should include the 4-byte separator prefix [0x05, 0xFF, 0xFF, 0xFF]
// followed by the bitmap data (64 bytes for 16x32).
// Returns one or more chunks ready to send to the device.
func EncodeText(bitmaps [][]byte, opts TextOptions) ([][]byte, error) {
	if len(bitmaps) == 0 {
		return nil, fmt.Errorf("no bitmaps provided")
	}

	numChars := len(bitmaps)

	// Build metadata (14 bytes)
	// Layout: [count:2][0x00][0x01][mode][speed][colormode][r][g][b][bgmode][bgr][bgg][bgb]
	metadata := make([]byte, textMetadataSize)
	binary.LittleEndian.PutUint16(metadata[0:2], uint16(numChars))
	metadata[2] = 0x00
	metadata[3] = 0x01
	metadata[4] = byte(opts.Mode)
	metadata[5] = opts.Speed
	metadata[6] = byte(opts.ColorMode)
	metadata[7] = opts.TextColor.R
	metadata[8] = opts.TextColor.G
	metadata[9] = opts.TextColor.B

	if opts.Background != nil {
		metadata[10] = 0x01
		metadata[11] = opts.Background.R
		metadata[12] = opts.Background.G
		metadata[13] = opts.Background.B
	}

	// Concatenate all bitmap data
	var allBitmaps []byte
	for _, bm := range bitmaps {
		allBitmaps = append(allBitmaps, bm...)
	}

	// payload = metadata + bitmaps
	payload := append(metadata, allBitmaps...)

	// CRC32 covers metadata + bitmaps
	crc := CRC32LE(payload)

	// Build header (16 bytes)
	totalLen := textHeaderSize + len(payload)
	header := make([]byte, textHeaderSize)
	binary.LittleEndian.PutUint16(header[0:2], uint16(totalLen))
	header[2] = 0x03 // opcode
	header[3] = 0x00
	header[4] = 0x00
	binary.LittleEndian.PutUint32(header[5:9], uint32(len(payload)))
	copy(header[9:13], crc[:])
	header[13] = 0x00
	header[14] = 0x00
	header[15] = 0x0C

	fullPacket := append(header, payload...)

	return [][]byte{fullPacket}, nil
}
