package main

import (
	"fmt"

	"github.com/jeffory/idotmatrix-lib/command"
	"github.com/spf13/cobra"
)

var gifCmd = &cobra.Command{
	Use:   "gif [file]",
	Short: "Upload a GIF animation to the device",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		dev, cleanup, err := connectDevice()
		if err != nil { return err }
		defer cleanup()
		c, err := command.NewGIFFromFile(args[0], getDisplaySize())
		if err != nil { return fmt.Errorf("load GIF: %w", err) }
		if err := dev.Send(c); err != nil { return fmt.Errorf("send: %w", err) }
		fmt.Printf("GIF uploaded: %s\n", args[0])
		return nil
	},
}

func init() { rootCmd.AddCommand(gifCmd) }
