package transport

import (
	"fmt"
	"strings"
	"time"

	"tinygo.org/x/bluetooth"
)

var adapter = bluetooth.DefaultAdapter

// DiscoveredDevice represents an iDotMatrix device found during scanning.
type DiscoveredDevice struct {
	Name    string
	Address string
	RSSI    int
}

// Discover scans for iDotMatrix devices for the given duration.
func Discover(timeout time.Duration) ([]DiscoveredDevice, error) {
	if err := adapter.Enable(); err != nil {
		return nil, fmt.Errorf("enable BLE adapter: %w", err)
	}

	var devices []DiscoveredDevice
	done := make(chan struct{})

	go func() {
		time.Sleep(timeout)
		adapter.StopScan()
		close(done)
	}()

	err := adapter.Scan(func(adapter *bluetooth.Adapter, result bluetooth.ScanResult) {
		name := result.LocalName()
		if strings.HasPrefix(name, DevicePrefix) {
			devices = append(devices, DiscoveredDevice{
				Name:    name,
				Address: result.Address.String(),
				RSSI:    int(result.RSSI),
			})
		}
	})
	if err != nil {
		return nil, fmt.Errorf("scan: %w", err)
	}

	<-done
	return devices, nil
}

// Connection manages a BLE connection to an iDotMatrix device.
type Connection struct {
	device    bluetooth.Device
	writeChar bluetooth.DeviceCharacteristic
	opts      ConnectOptions
	ackCh     chan []byte
}

// Connect establishes a BLE connection to an iDotMatrix device by address.
func Connect(address string, opts ConnectOptions) (*Connection, error) {
	if err := adapter.Enable(); err != nil {
		return nil, fmt.Errorf("enable BLE adapter: %w", err)
	}

	addr, err := bluetooth.ParseMAC(address)
	if err != nil {
		return nil, fmt.Errorf("parse address: %w", err)
	}

	device, err := adapter.Connect(bluetooth.Address{MACAddress: bluetooth.MACAddress{MAC: addr}}, bluetooth.ConnectionParams{})
	if err != nil {
		return nil, fmt.Errorf("connect: %w", err)
	}

	serviceUUID, _ := bluetooth.ParseUUID(ServiceUUID)
	services, err := device.DiscoverServices([]bluetooth.UUID{serviceUUID})
	if err != nil {
		device.Disconnect()
		return nil, fmt.Errorf("discover services: %w", err)
	}

	if len(services) == 0 {
		device.Disconnect()
		return nil, fmt.Errorf("iDotMatrix service not found")
	}

	writeUUID, _ := bluetooth.ParseUUID(WriteCharUUID)
	notifyUUID, _ := bluetooth.ParseUUID(NotifyCharUUID)

	chars, err := services[0].DiscoverCharacteristics([]bluetooth.UUID{writeUUID, notifyUUID})
	if err != nil {
		device.Disconnect()
		return nil, fmt.Errorf("discover characteristics: %w", err)
	}

	conn := &Connection{
		device: device,
		opts:   opts,
		ackCh:  make(chan []byte, 10),
	}

	for _, ch := range chars {
		if ch.UUID().String() == WriteCharUUID {
			conn.writeChar = ch
		}
		if ch.UUID().String() == NotifyCharUUID {
			ch.EnableNotifications(func(buf []byte) {
				// Copy buffer since it may be reused
				data := make([]byte, len(buf))
				copy(data, buf)
				select {
				case conn.ackCh <- data:
				default:
				}
			})
		}
	}

	return conn, nil
}

// Close disconnects from the device.
func (c *Connection) Close() error {
	return c.device.Disconnect()
}

// WritePacket sends a single command packet to the device.
func (c *Connection) WritePacket(data []byte) error {
	_, err := c.writeChar.WriteWithoutResponse(data)
	return err
}
