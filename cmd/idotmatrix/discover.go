package main

import (
	"fmt"
	"time"

	"github.com/jeffory/idotmatrix-lib/transport"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var saveDevice bool

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

		if saveDevice {
			addr := devices[0].Address
			viper.Set("device.address", addr)
			if err := saveConfig(); err != nil {
				return fmt.Errorf("save config: %w", err)
			}
			fmt.Printf("Saved %s as default device address.\n", addr)
		}

		return nil
	},
}

func init() {
	discoverCmd.Flags().BoolVar(&saveDevice, "save", false, "save first discovered device as default address")
	rootCmd.AddCommand(discoverCmd)
}
