package main

import (
	"fmt"
	"time"

	"github.com/jeffory/idotmatrix-lib/transport"
)

func main() {
	fmt.Println("Scanning for iDotMatrix devices (5 seconds)...")
	devices, err := transport.Discover(5 * time.Second)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	if len(devices) == 0 {
		fmt.Println("No devices found.")
		return
	}
	for _, d := range devices {
		fmt.Printf("  %s  %s  RSSI: %d\n", d.Name, d.Address, d.RSSI)
	}
}
