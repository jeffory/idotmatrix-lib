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
	cmd := command.NewText("Hello!",
		command.WithTextMode(protocol.TextScrollLeft),
		command.WithSpeed(128),
		command.WithColorMode(protocol.ColorFixed),
		command.WithTextColor(protocol.Color{R: 0, G: 255, B: 0}),
		command.WithDisplaySize(protocol.Display32x32),
	)
	if err := dev.Send(cmd); err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	fmt.Println("Text displayed: Hello!")
}
