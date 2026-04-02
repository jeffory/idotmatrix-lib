package main

import (
	"fmt"
	"os"
	"time"

	"github.com/jeffory/idotmatrix-lib/device"
	"github.com/jeffory/idotmatrix-lib/protocol"
	"github.com/jeffory/idotmatrix-lib/transport"
	"github.com/spf13/cobra"
)

var (
	address     string
	displaySize int
	timeout     int
)

var rootCmd = &cobra.Command{
	Use:   "idotmatrix",
	Short: "Control iDotMatrix pixel displays",
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&address, "address", "a", os.Getenv("IDOTMATRIX_ADDRESS"), "device MAC address (or set IDOTMATRIX_ADDRESS)")
	rootCmd.PersistentFlags().IntVarP(&displaySize, "size", "s", 32, "display size: 16, 32, or 64")
	rootCmd.PersistentFlags().IntVar(&timeout, "timeout", 10, "connection timeout in seconds")
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func getDisplaySize() protocol.DisplaySize {
	switch displaySize {
	case 16:
		return protocol.Display16x16
	case 64:
		return protocol.Display64x64
	default:
		return protocol.Display32x32
	}
}

func connectDevice() (*device.Device, func(), error) {
	if address == "" {
		return nil, nil, fmt.Errorf("device address required (use -a or set IDOTMATRIX_ADDRESS)")
	}
	opts := transport.ConnectOptions{
		Timeout:        time.Duration(timeout) * time.Second,
		ChunkDelay:     1 * time.Second,
		UseFlowControl: true,
	}
	conn, err := transport.Connect(address, opts)
	if err != nil {
		return nil, nil, fmt.Errorf("connect: %w", err)
	}
	dev := device.New(conn, getDisplaySize())
	cleanup := func() { conn.Close() }
	return dev, cleanup, nil
}
