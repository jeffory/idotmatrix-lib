package main

import (
	"fmt"
	"time"

	"github.com/jeffory/idotmatrix-lib/transport"
	"github.com/spf13/cobra"
)

var discoverCmd = &cobra.Command{
	Use:   "discover",
	Short: "Scan for nearby iDotMatrix devices",
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("Scanning for iDotMatrix devices...")
		devices, err := transport.Discover(time.Duration(timeout) * time.Second)
		if err != nil {
			return fmt.Errorf("scan failed: %w", err)
		}
		if len(devices) == 0 {
			fmt.Println("No devices found.")
			return nil
		}
		for _, d := range devices {
			fmt.Printf("  %s  %s  RSSI: %d\n", d.Name, d.Address, d.RSSI)
		}
		return nil
	},
}

func init() { rootCmd.AddCommand(discoverCmd) }
