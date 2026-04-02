package transport

import (
	"fmt"
	"time"
)

// WriteChunked sends multiple chunks with flow control.
// If UseFlowControl is true, waits for notification ACK between chunks.
// Otherwise, uses ChunkDelay between chunks.
func (c *Connection) WriteChunked(chunks [][]byte) error {
	for i, chunk := range chunks {
		err := c.WritePacket(chunk)
		if err != nil {
			return fmt.Errorf("write chunk %d: %w", i, err)
		}

		// Don't wait after the last chunk — just wait for completion
		if i < len(chunks)-1 {
			if c.opts.UseFlowControl {
				if err := c.waitForACK(c.opts.Timeout); err != nil {
					return fmt.Errorf("chunk %d ACK: %w", i, err)
				}
			} else {
				time.Sleep(c.opts.ChunkDelay)
			}
		}
	}

	// Wait for final completion ACK
	if c.opts.UseFlowControl {
		return c.waitForCompletion(c.opts.Timeout)
	}

	return nil
}

func (c *Connection) waitForACK(timeout time.Duration) error {
	select {
	case <-c.ackCh:
		return nil
	case <-time.After(timeout):
		return fmt.Errorf("timeout waiting for chunk ACK")
	}
}

func (c *Connection) waitForCompletion(timeout time.Duration) error {
	select {
	case ack := <-c.ackCh:
		if len(ack) >= 5 && ack[4] == 0x03 {
			return nil
		}
		return nil // Accept any ACK as completion
	case <-time.After(timeout):
		return fmt.Errorf("timeout waiting for upload completion")
	}
}
