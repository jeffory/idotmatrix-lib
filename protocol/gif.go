package protocol

import (
	"encoding/binary"
	"fmt"
)

const (
	gifHeaderSize = 16
	gifChunkSize  = 4096
)

// EncodeGIF splits raw GIF data into chunked packets with headers.
// Each chunk has a 16-byte header followed by up to 4096 bytes of GIF data.
// The CRC32 covers the raw GIF data only (not headers).
func EncodeGIF(data []byte, size DisplaySize) ([][]byte, error) {
	if len(data) == 0 {
		return nil, fmt.Errorf("empty GIF data")
	}

	crc := CRC32LE(data)

	// Split GIF data into chunks
	var gifChunks [][]byte
	for i := 0; i < len(data); i += gifChunkSize {
		end := i + gifChunkSize
		if end > len(data) {
			end = len(data)
		}
		gifChunks = append(gifChunks, data[i:end])
	}

	multiChunk := len(gifChunks) > 1
	totalPayload := uint32(len(data) + len(gifChunks)*gifHeaderSize)

	var result [][]byte
	for i, chunk := range gifChunks {
		header := make([]byte, gifHeaderSize)

		// Bytes 0-1: this chunk size including header (LE)
		binary.LittleEndian.PutUint16(header[0:2], uint16(len(chunk)+gifHeaderSize))

		header[2] = 0x01 // fixed
		header[3] = 0x00 // fixed

		// Byte 4: multi-chunk indicator
		// From Python reference: first chunk = 0x00, subsequent = 0x02
		if multiChunk && i > 0 {
			header[4] = 0x02
		} else {
			header[4] = 0x00
		}

		// Bytes 5-8: total payload across all chunks (LE)
		binary.LittleEndian.PutUint32(header[5:9], totalPayload)

		// Bytes 9-12: CRC32 of raw GIF (LE)
		copy(header[9:13], crc[:])

		// Bytes 13-15: fixed
		header[13] = 0x05
		header[14] = 0x00
		header[15] = 0x0D

		result = append(result, append(header, chunk...))
	}

	return result, nil
}
