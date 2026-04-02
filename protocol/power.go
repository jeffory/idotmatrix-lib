package protocol

// EncodePowerOn returns the packet to turn the display on.
func EncodePowerOn() []byte {
	return []byte{0x05, 0x00, 0x07, 0x01, 0x01}
}

// EncodePowerOff returns the packet to turn the display off.
func EncodePowerOff() []byte {
	return []byte{0x05, 0x00, 0x07, 0x01, 0x00}
}

// EncodeReset returns the two packets needed for a device reset.
func EncodeReset() (part1 []byte, part2 []byte) {
	return []byte{0x04, 0x00, 0x03, 0x80}, []byte{0x05, 0x00, 0x04, 0x80, 0x50}
}
