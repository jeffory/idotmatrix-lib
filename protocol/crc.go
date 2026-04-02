package protocol

import (
	"encoding/binary"
	"hash/crc32"
)

// CRC32LE computes a standard zlib CRC32 and returns 4 bytes in little-endian order.
func CRC32LE(data []byte) [4]byte {
	checksum := crc32.ChecksumIEEE(data)
	var result [4]byte
	binary.LittleEndian.PutUint32(result[:], checksum)
	return result
}
