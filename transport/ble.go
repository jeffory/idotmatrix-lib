package transport

import (
	"fmt"
	"strings"
	"time"

	"github.com/godbus/dbus/v5"
)

// DiscoveredDevice represents an iDotMatrix device found during scanning.
type DiscoveredDevice struct {
	Name    string
	Address string
	RSSI    int
}

// Discover scans for iDotMatrix devices for the given duration.
// It talks to BlueZ directly via DBus, which works even when another process
// (e.g., GNOME Bluetooth) already has discovery running.
func Discover(timeout time.Duration) ([]DiscoveredDevice, error) {
	bus, err := dbus.SystemBus()
	if err != nil {
		return nil, fmt.Errorf("connect to system bus: %w", err)
	}

	hci := bus.Object("org.bluez", "/org/bluez/hci0")

	// Set LE transport filter — iDotMatrix is a BLE device
	hci.Call("org.bluez.Adapter1.SetDiscoveryFilter", 0,
		map[string]dbus.Variant{
			"Transport": dbus.MakeVariant("le"),
		})

	// Listen for new devices appearing via DBus signals
	signal := make(chan *dbus.Signal, 50)
	bus.Signal(signal)
	defer bus.RemoveSignal(signal)

	bus.AddMatchSignal(
		dbus.WithMatchInterface("org.freedesktop.DBus.ObjectManager"),
		dbus.WithMatchMember("InterfacesAdded"),
	)
	bus.AddMatchSignal(
		dbus.WithMatchInterface("org.freedesktop.DBus.Properties"),
		dbus.WithMatchMember("PropertiesChanged"),
	)

	// Try to start discovery. Ignore "already in progress".
	call := hci.Call("org.bluez.Adapter1.StartDiscovery", 0)
	weStartedDiscovery := call.Err == nil

	seen := make(map[string]bool)
	var devices []DiscoveredDevice

	addDevice := func(name, addr string, rssi int) {
		if strings.HasPrefix(name, DevicePrefix) && !seen[addr] {
			seen[addr] = true
			devices = append(devices, DiscoveredDevice{
				Name:    name,
				Address: addr,
				RSSI:    rssi,
			})
			fmt.Printf("  Found: %s  %s  RSSI: %d\n", name, addr, rssi)
		}
	}

	// Check already-known devices first
	root := bus.Object("org.bluez", "/")
	var managedObjects map[dbus.ObjectPath]map[string]map[string]dbus.Variant
	if err := root.Call("org.freedesktop.DBus.ObjectManager.GetManagedObjects", 0).Store(&managedObjects); err == nil {
		for _, ifaces := range managedObjects {
			if dev, ok := ifaces["org.bluez.Device1"]; ok {
				name, _ := dev["Name"].Value().(string)
				addr, _ := dev["Address"].Value().(string)
				rssi, _ := dev["RSSI"].Value().(int16)
				addDevice(name, addr, int(rssi))
			}
		}
	}

	// Watch for newly discovered devices until timeout
	timer := time.After(timeout)
loop:
	for {
		select {
		case <-timer:
			break loop
		case sig := <-signal:
			switch sig.Name {
			case "org.freedesktop.DBus.ObjectManager.InterfacesAdded":
				if len(sig.Body) < 2 {
					continue
				}
				ifaces, ok := sig.Body[1].(map[string]map[string]dbus.Variant)
				if !ok {
					continue
				}
				if dev, ok := ifaces["org.bluez.Device1"]; ok {
					name, _ := dev["Name"].Value().(string)
					addr, _ := dev["Address"].Value().(string)
					rssi, _ := dev["RSSI"].Value().(int16)
					addDevice(name, addr, int(rssi))
				}
			case "org.freedesktop.DBus.Properties.PropertiesChanged":
				if len(sig.Body) < 2 {
					continue
				}
				iface, _ := sig.Body[0].(string)
				if iface != "org.bluez.Device1" {
					continue
				}
				changes, ok := sig.Body[1].(map[string]dbus.Variant)
				if !ok {
					continue
				}
				// A device just got a Name — check if it's ours
				if nameV, ok := changes["Name"]; ok {
					name, _ := nameV.Value().(string)
					// Get the address from the device path
					addrProp, err := bus.Object("org.bluez", sig.Path).GetProperty("org.bluez.Device1.Address")
					if err == nil {
						addr, _ := addrProp.Value().(string)
						addDevice(name, addr, 0)
					}
				}
			}
		}
	}

	if weStartedDiscovery {
		hci.Call("org.bluez.Adapter1.StopDiscovery", 0)
	}

	return devices, nil
}

// Connection manages a BLE connection to an iDotMatrix device via BlueZ DBus.
type Connection struct {
	bus       *dbus.Conn
	devPath   dbus.ObjectPath
	writePath dbus.ObjectPath
	opts      ConnectOptions
	ackCh     chan []byte
}

// findDevicePath finds the BlueZ DBus object path for a device by MAC address.
func findDevicePath(bus *dbus.Conn, address string) (dbus.ObjectPath, error) {
	upper := strings.ToUpper(address)

	root := bus.Object("org.bluez", "/")
	var managed map[dbus.ObjectPath]map[string]map[string]dbus.Variant
	if err := root.Call("org.freedesktop.DBus.ObjectManager.GetManagedObjects", 0).Store(&managed); err != nil {
		return "", fmt.Errorf("get managed objects: %w", err)
	}

	for path, ifaces := range managed {
		if dev, ok := ifaces["org.bluez.Device1"]; ok {
			addr, _ := dev["Address"].Value().(string)
			if strings.EqualFold(addr, upper) {
				return path, nil
			}
		}
	}
	return "", fmt.Errorf("device %s not found in BlueZ", address)
}

// ensureDeviceKnown makes sure BlueZ has the target device in its cache,
// triggering a BLE scan if necessary.
func ensureDeviceKnown(bus *dbus.Conn, address string, timeout time.Duration) error {
	upper := strings.ToUpper(address)

	// Check if BlueZ already knows this device.
	root := bus.Object("org.bluez", "/")
	var managed map[dbus.ObjectPath]map[string]map[string]dbus.Variant
	if err := root.Call("org.freedesktop.DBus.ObjectManager.GetManagedObjects", 0).Store(&managed); err == nil {
		for _, ifaces := range managed {
			if dev, ok := ifaces["org.bluez.Device1"]; ok {
				addr, _ := dev["Address"].Value().(string)
				if strings.EqualFold(addr, upper) {
					return nil
				}
			}
		}
	}

	// Not found — start a LE discovery scan.
	hci := bus.Object("org.bluez", "/org/bluez/hci0")
	hci.Call("org.bluez.Adapter1.SetDiscoveryFilter", 0,
		map[string]dbus.Variant{
			"Transport": dbus.MakeVariant("le"),
		})

	signal := make(chan *dbus.Signal, 50)
	bus.Signal(signal)
	defer bus.RemoveSignal(signal)

	bus.AddMatchSignal(
		dbus.WithMatchInterface("org.freedesktop.DBus.ObjectManager"),
		dbus.WithMatchMember("InterfacesAdded"),
	)

	call := hci.Call("org.bluez.Adapter1.StartDiscovery", 0)
	weStarted := call.Err == nil
	defer func() {
		if weStarted {
			hci.Call("org.bluez.Adapter1.StopDiscovery", 0)
		}
	}()

	fmt.Printf("Scanning for %s...\n", address)

	timer := time.After(timeout)
	for {
		select {
		case <-timer:
			return fmt.Errorf("device %s not found after %s scan", address, timeout)
		case sig := <-signal:
			if sig.Name != "org.freedesktop.DBus.ObjectManager.InterfacesAdded" || len(sig.Body) < 2 {
				continue
			}
			ifaces, ok := sig.Body[1].(map[string]map[string]dbus.Variant)
			if !ok {
				continue
			}
			if dev, ok := ifaces["org.bluez.Device1"]; ok {
				addr, _ := dev["Address"].Value().(string)
				if strings.EqualFold(addr, upper) {
					fmt.Println("Device found.")
					return nil
				}
			}
		}
	}
}

// waitServicesResolved waits for BlueZ to finish GATT service discovery.
func waitServicesResolved(bus *dbus.Conn, devPath dbus.ObjectPath, timeout time.Duration) error {
	dev := bus.Object("org.bluez", devPath)

	// Check if already resolved.
	prop, err := dev.GetProperty("org.bluez.Device1.ServicesResolved")
	if err == nil {
		if resolved, ok := prop.Value().(bool); ok && resolved {
			return nil
		}
	}

	// Watch for the property to change.
	signal := make(chan *dbus.Signal, 10)
	bus.Signal(signal)
	defer bus.RemoveSignal(signal)

	bus.AddMatchSignal(
		dbus.WithMatchInterface("org.freedesktop.DBus.Properties"),
		dbus.WithMatchMember("PropertiesChanged"),
		dbus.WithMatchObjectPath(devPath),
	)

	timer := time.After(timeout)
	for {
		select {
		case <-timer:
			return fmt.Errorf("timeout waiting for GATT services to resolve")
		case sig := <-signal:
			if sig.Path != devPath || len(sig.Body) < 2 {
				continue
			}
			changes, ok := sig.Body[1].(map[string]dbus.Variant)
			if !ok {
				continue
			}
			if v, ok := changes["ServicesResolved"]; ok {
				if resolved, ok := v.Value().(bool); ok && resolved {
					return nil
				}
			}
		}
	}
}

// findCharacteristic finds a GATT characteristic object path by UUID under a device.
func findCharacteristic(bus *dbus.Conn, devPath dbus.ObjectPath, charUUID string) (dbus.ObjectPath, error) {
	root := bus.Object("org.bluez", "/")
	var managed map[dbus.ObjectPath]map[string]map[string]dbus.Variant
	if err := root.Call("org.freedesktop.DBus.ObjectManager.GetManagedObjects", 0).Store(&managed); err != nil {
		return "", fmt.Errorf("get managed objects: %w", err)
	}

	prefix := string(devPath) + "/"
	for path, ifaces := range managed {
		if !strings.HasPrefix(string(path), prefix) {
			continue
		}
		if char, ok := ifaces["org.bluez.GattCharacteristic1"]; ok {
			uuid, _ := char["UUID"].Value().(string)
			if strings.EqualFold(uuid, charUUID) {
				return path, nil
			}
		}
	}
	return "", fmt.Errorf("characteristic %s not found", charUUID)
}

// Connect establishes a BLE connection to an iDotMatrix device by address.
func Connect(address string, opts ConnectOptions) (*Connection, error) {
	bus, err := dbus.SystemBus()
	if err != nil {
		return nil, fmt.Errorf("connect to system bus: %w", err)
	}

	if err := ensureDeviceKnown(bus, address, opts.Timeout); err != nil {
		return nil, fmt.Errorf("scan for device: %w", err)
	}

	devPath, err := findDevicePath(bus, address)
	if err != nil {
		return nil, err
	}

	// Connect via BlueZ.
	dev := bus.Object("org.bluez", devPath)
	if call := dev.Call("org.bluez.Device1.Connect", 0); call.Err != nil {
		return nil, fmt.Errorf("connect: %w", call.Err)
	}

	// Wait for GATT services to be discovered.
	if err := waitServicesResolved(bus, devPath, opts.Timeout); err != nil {
		dev.Call("org.bluez.Device1.Disconnect", 0)
		return nil, err
	}

	// Find the write characteristic.
	writePath, err := findCharacteristic(bus, devPath, WriteCharUUID)
	if err != nil {
		dev.Call("org.bluez.Device1.Disconnect", 0)
		return nil, fmt.Errorf("find write characteristic: %w", err)
	}

	// Find the notify characteristic and start notifications.
	notifyPath, err := findCharacteristic(bus, devPath, NotifyCharUUID)
	if err != nil {
		dev.Call("org.bluez.Device1.Disconnect", 0)
		return nil, fmt.Errorf("find notify characteristic: %w", err)
	}

	conn := &Connection{
		bus:       bus,
		devPath:   devPath,
		writePath: writePath,
		opts:      opts,
		ackCh:     make(chan []byte, 10),
	}

	// Subscribe to value changes on the notify characteristic.
	bus.AddMatchSignal(
		dbus.WithMatchInterface("org.freedesktop.DBus.Properties"),
		dbus.WithMatchMember("PropertiesChanged"),
		dbus.WithMatchObjectPath(notifyPath),
	)

	signal := make(chan *dbus.Signal, 50)
	bus.Signal(signal)

	go func() {
		for sig := range signal {
			if len(sig.Body) < 2 {
				continue
			}
			iface, _ := sig.Body[0].(string)
			if iface != "org.bluez.GattCharacteristic1" {
				continue
			}
			changes, ok := sig.Body[1].(map[string]dbus.Variant)
			if !ok {
				continue
			}
			if v, ok := changes["Value"]; ok {
				if buf, ok := v.Value().([]byte); ok {
					data := make([]byte, len(buf))
					copy(data, buf)
					select {
					case conn.ackCh <- data:
					default:
					}
				}
			}
		}
	}()

	// Tell BlueZ to start sending notifications.
	notifyObj := bus.Object("org.bluez", notifyPath)
	if call := notifyObj.Call("org.bluez.GattCharacteristic1.StartNotify", 0); call.Err != nil {
		dev.Call("org.bluez.Device1.Disconnect", 0)
		return nil, fmt.Errorf("start notifications: %w", call.Err)
	}

	return conn, nil
}

// Close disconnects from the device.
func (c *Connection) Close() error {
	dev := c.bus.Object("org.bluez", c.devPath)
	return dev.Call("org.bluez.Device1.Disconnect", 0).Err
}

// WritePacket sends a single command packet to the device.
// Packets larger than 20 bytes use write-with-response ("request") so BlueZ
// handles ATT segmentation. Small packets use write-without-response ("command")
// for lower latency.
func (c *Connection) WritePacket(data []byte) error {
	writeType := "command"
	if len(data) > 20 {
		writeType = "request"
	}
	char := c.bus.Object("org.bluez", c.writePath)
	call := char.Call("org.bluez.GattCharacteristic1.WriteValue", 0, data, map[string]dbus.Variant{
		"type": dbus.MakeVariant(writeType),
	})
	return call.Err
}
