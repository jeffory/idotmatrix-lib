package main

import (
	"fmt"
	"os"

	"github.com/jeffory/idotmatrix-lib/command"
	"github.com/jeffory/idotmatrix-lib/device"
	"github.com/jeffory/idotmatrix-lib/protocol"
	"github.com/jeffory/idotmatrix-lib/transport"
)

func main() {
	addr := os.Getenv("IDOTMATRIX_ADDRESS")
	if addr == "" {
		fmt.Println("Set IDOTMATRIX_ADDRESS env var to device MAC")
		return
	}
	conn, err := transport.Connect(addr, transport.DefaultOptions())
	if err != nil {
		fmt.Printf("Connect error: %v\n", err)
		return
	}
	defer conn.Close()

	dev := device.New(conn, protocol.Display32x32)
	fmt.Println("Turning on...")
	if err := dev.Send(command.PowerOn()); err != nil {
		fmt.Printf("Error: %v\n", err)
	}
}
