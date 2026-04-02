package main

import (
	"fmt"

	"github.com/jeffory/idotmatrix-lib/command"
	"github.com/jeffory/idotmatrix-lib/protocol"
	"github.com/spf13/cobra"
)

var (
	textMode  string
	textSpeed int
	textColor string
)

var textCmd = &cobra.Command{
	Use:   "text [message]",
	Short: "Display text on the device",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		dev, cleanup, err := connectDevice()
		if err != nil { return err }
		defer cleanup()
		color, err := parseHexColor(textColor)
		if err != nil { return fmt.Errorf("invalid color: %w", err) }
		c := command.NewText(args[0],
			command.WithTextMode(parseTextMode(textMode)),
			command.WithSpeed(uint8(textSpeed)),
			command.WithTextColor(color),
			command.WithDisplaySize(getDisplaySize()),
		)
		if err := dev.Send(c); err != nil { return fmt.Errorf("send: %w", err) }
		fmt.Printf("Text displayed: %q\n", args[0])
		return nil
	},
}

func init() {
	textCmd.Flags().StringVar(&textMode, "mode", "fixed", "text mode: fixed, scroll-left, scroll-right, scroll-up, scroll-down, strobe, fade, falling, laser")
	textCmd.Flags().IntVar(&textSpeed, "speed", 100, "animation speed (0-255)")
	textCmd.Flags().StringVar(&textColor, "color", "ffffff", "hex color (e.g. ff0000)")
	rootCmd.AddCommand(textCmd)
}

func parseTextMode(s string) protocol.TextMode {
	switch s {
	case "scroll-left":  return protocol.TextScrollLeft
	case "scroll-right": return protocol.TextScrollRight
	case "scroll-up":    return protocol.TextScrollUp
	case "scroll-down":  return protocol.TextScrollDown
	case "strobe":       return protocol.TextStrobe
	case "fade":         return protocol.TextFade
	case "falling":      return protocol.TextFalling
	case "laser":        return protocol.TextLaser
	default:             return protocol.TextFixed
	}
}
