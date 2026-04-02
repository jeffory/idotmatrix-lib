package main

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"time"

	"github.com/jeffory/idotmatrix-lib/command"
	"github.com/jeffory/idotmatrix-lib/pixellab"
	"github.com/spf13/cobra"
)

var (
	animAPIKey    string
	animFrames    int
	animDelay     int
	animView      string
	animDirection string
	animGuidance  float64
	animSavePath  string
	animPalette   []string
)

var animateCmd = &cobra.Command{
	Use:   "animate [description] [action]",
	Short: "Generate a pixel art animation with AI and display it",
	Long:  "Generate a pixel art animation using the Pixellab API and send it to the display.",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		if animAPIKey == "" {
			return fmt.Errorf("API key required (use --api-key or set PIXELLAB_API_KEY)")
		}

		client := pixellab.New(animAPIKey)
		imgSize := pixellab.ImageSize{Width: 64, Height: 64} // API only supports 64x64

		ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
		defer cancel()

		// Step 1: Generate a reference image from the description.
		fmt.Println("Generating reference image...")
		refImg, err := client.Generate(ctx, pixellab.Request{
			Description: args[0],
			ImageSize:   imgSize,
		})
		if err != nil {
			return fmt.Errorf("generate reference image: %w", err)
		}

		refB64, err := pixellab.EncodeBase64Image(refImg)
		if err != nil {
			return fmt.Errorf("encode reference image: %w", err)
		}

		// Step 2: Animate using the reference image.
		req := pixellab.AnimationRequest{
			Description:    args[0],
			Action:         args[1],
			ImageSize:      imgSize,
			ReferenceImage: refB64,
		}
		if cmd.Flags().Changed("view") {
			req.View = animView
		}
		if cmd.Flags().Changed("direction") {
			req.Direction = animDirection
		}
		if cmd.Flags().Changed("frames") {
			req.NFrames = &animFrames
		}
		if cmd.Flags().Changed("guidance") {
			req.GuidanceScale = &animGuidance
		}
		if len(animPalette) > 0 {
			req.ForcedPalette = animPalette
		}

		fmt.Println("Generating animation...")
		frames, err := client.GenerateAnimation(ctx, req)
		if err != nil {
			return fmt.Errorf("generate animation: %w", err)
		}
		fmt.Printf("Got %d frames\n", len(frames))

		size := getDisplaySize()
		gifBytes, err := pixellab.FramesToGIF(frames, animDelay, &pixellab.ImageSize{
			Width:  size.Width,
			Height: size.Height,
		})
		if err != nil {
			return fmt.Errorf("encode GIF: %w", err)
		}

		if animSavePath != "" {
			if err := os.WriteFile(animSavePath, gifBytes, 0644); err != nil {
				return fmt.Errorf("save: %w", err)
			}
			fmt.Printf("Saved to %s\n", animSavePath)
		}

		dev, cleanup, err := connectDevice()
		if err != nil {
			return err
		}
		defer cleanup()

		c, err := command.NewGIF(bytes.NewReader(gifBytes), size)
		if err != nil {
			return fmt.Errorf("prepare GIF: %w", err)
		}
		if err := dev.Send(c); err != nil {
			return fmt.Errorf("send: %w", err)
		}

		fmt.Println("Animation displayed")
		return nil
	},
}

func init() {
	animateCmd.Flags().StringVar(&animAPIKey, "api-key", os.Getenv("PIXELLAB_API_KEY"), "Pixellab API key (or set PIXELLAB_API_KEY)")
	animateCmd.Flags().IntVar(&animFrames, "frames", 4, "number of animation frames")
	animateCmd.Flags().IntVar(&animDelay, "delay", 10, "frame delay in centiseconds (10 = 100ms)")
	animateCmd.Flags().StringVar(&animView, "view", "side", "camera view: side, low top-down, high top-down")
	animateCmd.Flags().StringVar(&animDirection, "direction", "south", "facing direction: south, north, east, west, south-east, south-west, north-east, north-west")
	animateCmd.Flags().Float64Var(&animGuidance, "guidance", 4.0, "guidance scale (1.0-20.0)")
	animateCmd.Flags().StringVar(&animSavePath, "save", "", "save generated GIF to file path")
	animateCmd.Flags().StringSliceVar(&animPalette, "palette", nil, "forced color palette (hex codes, comma-separated)")
	rootCmd.AddCommand(animateCmd)
}
