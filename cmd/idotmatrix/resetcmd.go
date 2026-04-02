package main

import (
	"fmt"

	"github.com/jeffory/idotmatrix-lib/command"
	"github.com/spf13/cobra"
)

var resetCmd = &cobra.Command{
	Use:   "reset",
	Short: "Reset the device",
	RunE: func(cmd *cobra.Command, args []string) error {
		dev, cleanup, err := connectDevice()
		if err != nil { return err }
		defer cleanup()
		if err := dev.Send(command.Reset()); err != nil { return fmt.Errorf("send: %w", err) }
		fmt.Println("Device reset")
		return nil
	},
}

func init() { rootCmd.AddCommand(resetCmd) }
