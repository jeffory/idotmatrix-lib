package transport

import "time"

const (
	ServiceUUID      = "000000fa-0000-1000-8000-00805f9b34fb"
	WriteCharUUID    = "0000fa02-0000-1000-8000-00805f9b34fb"
	NotifyCharUUID   = "0000fa03-0000-1000-8000-00805f9b34fb"
	DevicePrefix     = "IDM-"
	DefaultChunkSize = 4096
)

// ConnectOptions configures connection behavior.
type ConnectOptions struct {
	Timeout        time.Duration // connection timeout (default 10s)
	ChunkDelay     time.Duration // delay between chunks when flow control disabled (default 1s)
	UseFlowControl bool          // wait for notification ACKs between chunks (default true)
}

// DefaultOptions returns sensible default connection options.
func DefaultOptions() ConnectOptions {
	return ConnectOptions{
		Timeout:        10 * time.Second,
		ChunkDelay:     1 * time.Second,
		UseFlowControl: true,
	}
}
