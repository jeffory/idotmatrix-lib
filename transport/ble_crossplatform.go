//go:build darwin || windows

package transport

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"tinygo.org/x/bluetooth"
)

// Connection manages a BLE connection to an iDotMatrix device via tinygo-org/bluetooth.
type Connection struct {
	device    bluetooth.Device
	writeChar bluetooth.DeviceCharacteristic
	opts      ConnectOptions
	ackCh     chan []byte
}

var (
	adapter     = bluetooth.DefaultAdapter
	adapterOnce sync.Once
	adapterErr  error
)

func enableAdapter() error {
	adapterOnce.Do(func() {
		adapterErr = adapter.Enable()
	})
	return adapterErr
}

// Discover scans for iDotMatrix devices for the given duration.
func Discover(timeout time.Duration) ([]DiscoveredDevice, error) {
	if err := enableAdapter(); err != nil {
		return nil, fmt.Errorf("enable BLE adapter: %w", err)
	}

	var mu sync.Mutex
	var devices []DiscoveredDevice
	seen := make(map[string]bool)

	time.AfterFunc(timeout, func() {
		adapter.StopScan()
	})

	err := adapter.Scan(func(a *bluetooth.Adapter, result bluetooth.ScanResult) {
		addr := result.Address.String()
		name := result.LocalName()
		mu.Lock()
		defer mu.Unlock()
		if strings.HasPrefix(name, DevicePrefix) && !seen[addr] {
			seen[addr] = true
			devices = append(devices, DiscoveredDevice{
				Name:    name,
				Address: addr,
				RSSI:    int(result.RSSI),
			})
			fmt.Printf("  Found: %s  %s  RSSI: %d\n", name, addr, result.RSSI)
		}
	})
	if err != nil {
		return nil, fmt.Errorf("scan: %w", err)
	}

	return devices, nil
}

// Connect establishes a BLE connection to an iDotMatrix device by address.
// The address should match the string returned by Discover (MAC on Windows/Linux,
// UUID on macOS since CoreBluetooth does not expose MAC addresses).
func Connect(address string, opts ConnectOptions) (*Connection, error) {
	if err := enableAdapter(); err != nil {
		return nil, fmt.Errorf("enable BLE adapter: %w", err)
	}

	// Scan to find the device. We match by address string or device name,
	// since on macOS the address is a UUID rather than a MAC.
	var targetAddr bluetooth.Address
	found := make(chan struct{}, 1)

	time.AfterFunc(opts.Timeout, func() {
		adapter.StopScan()
	})

	err := adapter.Scan(func(a *bluetooth.Adapter, result bluetooth.ScanResult) {
		if strings.EqualFold(result.Address.String(), address) ||
			strings.EqualFold(result.LocalName(), address) {
			targetAddr = result.Address
			a.StopScan()
			select {
			case found <- struct{}{}:
			default:
			}
		}
	})

	select {
	case <-found:
	default:
		if err != nil {
			return nil, fmt.Errorf("scan: %w", err)
		}
		return nil, fmt.Errorf("device %s not found", address)
	}

	fmt.Printf("Connecting to %s...\n", address)

	device, err := adapter.Connect(targetAddr, bluetooth.ConnectionParams{})
	if err != nil {
		return nil, fmt.Errorf("connect: %w", err)
	}

	// Discover the iDotMatrix service.
	serviceUUID, err := bluetooth.ParseUUID(ServiceUUID)
	if err != nil {
		device.Disconnect()
		return nil, fmt.Errorf("parse service UUID: %w", err)
	}

	services, err := device.DiscoverServices([]bluetooth.UUID{serviceUUID})
	if err != nil {
		device.Disconnect()
		return nil, fmt.Errorf("discover services: %w", err)
	}
	if len(services) == 0 {
		device.Disconnect()
		return nil, fmt.Errorf("service %s not found", ServiceUUID)
	}

	// Discover write and notify characteristics.
	writeUUID, err := bluetooth.ParseUUID(WriteCharUUID)
	if err != nil {
		device.Disconnect()
		return nil, fmt.Errorf("parse write UUID: %w", err)
	}
	notifyUUID, err := bluetooth.ParseUUID(NotifyCharUUID)
	if err != nil {
		device.Disconnect()
		return nil, fmt.Errorf("parse notify UUID: %w", err)
	}

	chars, err := services[0].DiscoverCharacteristics([]bluetooth.UUID{writeUUID, notifyUUID})
	if err != nil {
		device.Disconnect()
		return nil, fmt.Errorf("discover characteristics: %w", err)
	}

	var writeChar, notifyChar bluetooth.DeviceCharacteristic
	var foundWrite, foundNotify bool
	for _, ch := range chars {
		switch ch.UUID() {
		case writeUUID:
			writeChar = ch
			foundWrite = true
		case notifyUUID:
			notifyChar = ch
			foundNotify = true
		}
	}
	if !foundWrite {
		device.Disconnect()
		return nil, fmt.Errorf("write characteristic %s not found", WriteCharUUID)
	}
	if !foundNotify {
		device.Disconnect()
		return nil, fmt.Errorf("notify characteristic %s not found", NotifyCharUUID)
	}

	conn := &Connection{
		device:    device,
		writeChar: writeChar,
		opts:      opts,
		ackCh:     make(chan []byte, 10),
	}

	// Subscribe to notifications.
	err = notifyChar.EnableNotifications(func(buf []byte) {
		data := make([]byte, len(buf))
		copy(data, buf)
		select {
		case conn.ackCh <- data:
		default:
		}
	})
	if err != nil {
		device.Disconnect()
		return nil, fmt.Errorf("enable notifications: %w", err)
	}

	return conn, nil
}

// Close disconnects from the device.
func (c *Connection) Close() error {
	return c.device.Disconnect()
}

// WritePacket sends a single command packet to the device.
// Packets larger than 20 bytes use write-with-response so the BLE stack
// handles ATT segmentation. Small packets use write-without-response
// for lower latency.
func (c *Connection) WritePacket(data []byte) error {
	if len(data) > 20 {
		_, err := c.writeChar.Write(data)
		return err
	}
	_, err := c.writeChar.WriteWithoutResponse(data)
	return err
}
