package main

import (
	"context"
	"fmt"
	"image/png"
	"os"
	"time"

	"github.com/jeffory/idotmatrix-lib/command"
	"github.com/jeffory/idotmatrix-lib/pixellab"
	"github.com/spf13/cobra"
)

var (
	apiKey   string
	negative string
	guidance float64
	noBg     bool
	outline  string
	shading  string
	detail   string
	seed     int
	savePath string
)

var generateCmd = &cobra.Command{
	Use:   "generate [description]",
	Short: "Generate pixel art with AI and display it",
	Long:  "Generate pixel art using the Pixellab API and send it to the display.",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if apiKey == "" {
			return fmt.Errorf("API key required (use --api-key or set PIXELLAB_API_KEY)")
		}

		size := getDisplaySize()
		client := pixellab.New(apiKey)

		req := pixellab.Request{
			Description: args[0],
			ImageSize:   pixellab.ImageSize{Width: size.Width, Height: size.Height},
		}
		if cmd.Flags().Changed("negative") {
			req.NegativeDescription = &negative
		}
		if cmd.Flags().Changed("guidance") {
			req.TextGuidanceScale = &guidance
		}
		if cmd.Flags().Changed("no-background") {
			req.NoBackground = &noBg
		}
		if cmd.Flags().Changed("outline") {
			req.Outline = &outline
		}
		if cmd.Flags().Changed("shading") {
			req.Shading = &shading
		}
		if cmd.Flags().Changed("detail") {
			req.Detail = &detail
		}
		if cmd.Flags().Changed("seed") {
			req.Seed = &seed
		}

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		fmt.Println("Generating image...")
		img, err := client.Generate(ctx, req)
		if err != nil {
			return fmt.Errorf("generate: %w", err)
		}

		if savePath != "" {
			f, err := os.Create(savePath)
			if err != nil {
				return fmt.Errorf("save: %w", err)
			}
			if err := png.Encode(f, img); err != nil {
				f.Close()
				return fmt.Errorf("save: %w", err)
			}
			f.Close()
			fmt.Printf("Saved to %s\n", savePath)
		}

		dev, cleanup, err := connectDevice()
		if err != nil {
			return err
		}
		defer cleanup()

		c, err := command.NewImage(img, size, imageOpts()...)
		if err != nil {
			return fmt.Errorf("prepare image: %w", err)
		}
		if err := dev.Send(c); err != nil {
			return fmt.Errorf("send: %w", err)
		}

		fmt.Println("Generated image displayed")
		return nil
	},
}

func init() {
	generateCmd.Flags().StringVar(&apiKey, "api-key", os.Getenv("PIXELLAB_API_KEY"), "Pixellab API key (or set PIXELLAB_API_KEY)")
	generateCmd.Flags().StringVar(&negative, "negative", "", "negative description (what to avoid)")
	generateCmd.Flags().Float64Var(&guidance, "guidance", 8.0, "text guidance scale (1.0-20.0)")
	generateCmd.Flags().BoolVar(&noBg, "no-background", false, "generate with transparent background")
	generateCmd.Flags().StringVar(&outline, "outline", "", "outline style")
	generateCmd.Flags().StringVar(&shading, "shading", "", "shading style")
	generateCmd.Flags().StringVar(&detail, "detail", "", "detail level")
	generateCmd.Flags().IntVar(&seed, "seed", -1, "random seed (-1 = random)")
	generateCmd.Flags().StringVar(&savePath, "save", "", "save generated PNG to file path")
	rootCmd.AddCommand(generateCmd)
}
