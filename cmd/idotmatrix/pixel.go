package main

import (
	"fmt"
	"strconv"

	"github.com/jeffory/idotmatrix-lib/command"
	"github.com/jeffory/idotmatrix-lib/protocol"
	"github.com/spf13/cobra"
)

var (
	pixelX     int
	pixelY     int
	pixelColor string
)

var pixelCmd = &cobra.Command{
	Use:   "pixel",
	Short: "Draw a single pixel",
	RunE: func(cmd *cobra.Command, args []string) error {
		dev, cleanup, err := connectDevice()
		if err != nil { return err }
		defer cleanup()
		color, err := parseHexColor(pixelColor)
		if err != nil { return fmt.Errorf("invalid color: %w", err) }
		if err := dev.Send(command.DrawPixel(uint8(pixelX), uint8(pixelY), color)); err != nil {
			return fmt.Errorf("send: %w", err)
		}
		fmt.Printf("Pixel set at (%d,%d) color #%s\n", pixelX, pixelY, pixelColor)
		return nil
	},
}

func init() {
	pixelCmd.Flags().IntVar(&pixelX, "x", 0, "X coordinate")
	pixelCmd.Flags().IntVar(&pixelY, "y", 0, "Y coordinate")
	pixelCmd.Flags().StringVar(&pixelColor, "color", "ff0000", "hex color (e.g. ff0000)")
	rootCmd.AddCommand(pixelCmd)
}

func parseHexColor(s string) (protocol.Color, error) {
	if len(s) != 6 {
		return protocol.Color{}, fmt.Errorf("color must be 6 hex chars, got %q", s)
	}
	r, err := strconv.ParseUint(s[0:2], 16, 8)
	if err != nil { return protocol.Color{}, err }
	g, err := strconv.ParseUint(s[2:4], 16, 8)
	if err != nil { return protocol.Color{}, err }
	b, err := strconv.ParseUint(s[4:6], 16, 8)
	if err != nil { return protocol.Color{}, err }
	return protocol.Color{R: uint8(r), G: uint8(g), B: uint8(b)}, nil
}
