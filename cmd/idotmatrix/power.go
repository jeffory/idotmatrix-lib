package main

import (
	"fmt"

	"github.com/jeffory/idotmatrix-lib/command"
	"github.com/spf13/cobra"
)

var powerCmd = &cobra.Command{
	Use:   "power [on|off]",
	Short: "Turn the display on or off",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		dev, cleanup, err := connectDevice()
		if err != nil { return err }
		defer cleanup()
		var c command.Command
		switch args[0] {
		case "on":
			c = command.PowerOn()
		case "off":
			c = command.PowerOff()
		default:
			return fmt.Errorf("invalid argument: %s (use 'on' or 'off')", args[0])
		}
		if err := dev.Send(c); err != nil { return fmt.Errorf("send: %w", err) }
		fmt.Printf("Power %s\n", args[0])
		return nil
	},
}

func init() { rootCmd.AddCommand(powerCmd) }
