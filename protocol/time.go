package protocol

import "time"

// EncodeTimeSync returns the 11-byte packet to sync the device clock.
// Year is masked to a single byte (year & 0xFF).
// Day of week: Monday=1 through Sunday=7.
func EncodeTimeSync(t time.Time) []byte {
	year := byte(t.Year() & 0xFF)
	month := byte(t.Month())
	day := byte(t.Day())

	// Go's time.Weekday(): Sunday=0, Monday=1, ..., Saturday=6
	// Device expects: Monday=1, ..., Sunday=7
	dow := byte(t.Weekday())
	if dow == 0 {
		dow = 7 // Sunday
	}

	hour := byte(t.Hour())
	minute := byte(t.Minute())
	second := byte(t.Second())

	return []byte{
		0x0B, 0x00, // length (11 bytes, LE)
		0x01,       // opcode
		0x80,       // MIN_BYTE_VALUE (fixed)
		year, month, day, dow,
		hour, minute, second,
	}
}
