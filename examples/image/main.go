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
	if len(os.Args) < 2 {
		fmt.Println("Usage: image <path-to-image>")
		return
	}
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
	cmd, err := command.NewImageFromFile(os.Args[1], protocol.Display32x32)
	if err != nil {
		fmt.Printf("Load image error: %v\n", err)
		return
	}
	if err := dev.Send(cmd); err != nil {
		fmt.Printf("Send error: %v\n", err)
		return
	}
	fmt.Printf("Image displayed: %s\n", os.Args[1])
}
