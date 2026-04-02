package main

import (
	"fmt"
	"time"

	"github.com/jeffory/idotmatrix-lib/command"
	"github.com/spf13/cobra"
)

var timeCmd = &cobra.Command{
	Use:   "time",
	Short: "Sync the device clock with system time",
	RunE: func(cmd *cobra.Command, args []string) error {
		dev, cleanup, err := connectDevice()
		if err != nil { return err }
		defer cleanup()
		if err := dev.Send(command.SyncTime(time.Now())); err != nil { return fmt.Errorf("send: %w", err) }
		fmt.Println("Time synced")
		return nil
	},
}

func init() { rootCmd.AddCommand(timeCmd) }
