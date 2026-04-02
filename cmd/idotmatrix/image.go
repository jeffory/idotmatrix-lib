package main

import (
	"fmt"

	"github.com/jeffory/idotmatrix-lib/command"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var imageCmd = &cobra.Command{
	Use:   "image [file]",
	Short: "Display a static image on the device",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		dev, cleanup, err := connectDevice()
		if err != nil {
			return err
		}
		defer cleanup()
		c, err := command.NewImageFromFile(args[0], getDisplaySize(), imageOpts()...)
		if err != nil {
			return fmt.Errorf("load image: %w", err)
		}
		if err := dev.Send(c); err != nil {
			return fmt.Errorf("send: %w", err)
		}
		fmt.Printf("Image displayed: %s\n", args[0])
		return nil
	},
}

func init() { rootCmd.AddCommand(imageCmd) }

// imageOpts returns ImageOption values from flags/config.
func imageOpts() []command.ImageOption {
	var opts []command.ImageOption
	if v := viper.GetFloat64("image.contrast"); v != 0 {
		opts = append(opts, command.WithContrast(v))
	}
	if v := viper.GetFloat64("image.saturation"); v != 0 {
		opts = append(opts, command.WithSaturation(v))
	}
	return opts
}
