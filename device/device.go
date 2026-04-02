package device

import (
	"fmt"

	"github.com/jeffory/idotmatrix-lib/command"
	"github.com/jeffory/idotmatrix-lib/protocol"
	"github.com/jeffory/idotmatrix-lib/transport"
)

// Writer abstracts the transport layer for testability.
type Writer interface {
	WritePacket(data []byte) error
	WriteChunked(chunks [][]byte) error
	Close() error
}

// Device represents a connected iDotMatrix display.
type Device struct {
	writer Writer
	size   protocol.DisplaySize
}

// New creates a Device from a real BLE connection.
func New(conn *transport.Connection, size protocol.DisplaySize) *Device {
	return &Device{writer: conn, size: size}
}

// NewWithWriter creates a Device with a custom writer (for testing).
func NewWithWriter(w Writer, size protocol.DisplaySize) *Device {
	return &Device{writer: w, size: size}
}

// Send encodes and transmits a command to the device.
func (d *Device) Send(cmd command.Command) error {
	packets, err := cmd.Encode()
	if err != nil {
		return fmt.Errorf("encode command: %w", err)
	}

	if cmd.Chunked() {
		return d.writer.WriteChunked(packets)
	}

	for _, p := range packets {
		if err := d.writer.WritePacket(p); err != nil {
			return fmt.Errorf("write packet: %w", err)
		}
	}
	return nil
}

// Close disconnects from the device.
func (d *Device) Close() error {
	return d.writer.Close()
}
