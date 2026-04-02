package main

import (
	"fmt"

	"github.com/jeffory/idotmatrix-lib/command"
	"github.com/spf13/cobra"
)

var imageCmd = &cobra.Command{
	Use:   "image [file]",
	Short: "Display a static image on the device",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		dev, cleanup, err := connectDevice()
		if err != nil { return err }
		defer cleanup()
		c, err := command.NewImageFromFile(args[0], getDisplaySize())
		if err != nil { return fmt.Errorf("load image: %w", err) }
		if err := dev.Send(c); err != nil { return fmt.Errorf("send: %w", err) }
		fmt.Printf("Image displayed: %s\n", args[0])
		return nil
	},
}

func init() { rootCmd.AddCommand(imageCmd) }
