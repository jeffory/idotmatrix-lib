package command

import (
	"bytes"
	"image"
	"image/color"
	"image/draw"
	"image/gif"

	"github.com/jeffory/idotmatrix-lib/font"
	"github.com/jeffory/idotmatrix-lib/protocol"
)

// ScrollTextOption configures a scrolling text command.
type ScrollTextOption func(*scrollTextCmd)

type scrollTextCmd struct {
	text        string
	textColor   protocol.Color
	font        font.Renderer
	displaySize protocol.DisplaySize
	pixelsPerFrame int
	frameDelay     int // in 100ths of a second
}

// NewScrollText creates a command that renders text as a smoothly scrolling
// animated GIF. The text is drawn into one wide image, then panned across
// the display one pixel at a time.
func NewScrollText(text string, opts ...ScrollTextOption) Command {
	cmd := &scrollTextCmd{
		text:           text,
		textColor:      protocol.Color{R: 255, G: 255, B: 255},
		displaySize:    protocol.Display32x32,
		font:           font.DefaultBitmap(),
		pixelsPerFrame: 2,
		frameDelay:     8, // 80ms per frame = ~12fps
	}
	for _, opt := range opts {
		opt(cmd)
	}
	return cmd
}

func (*scrollTextCmd) Chunked() bool { return true }

func (c *scrollTextCmd) Encode() ([][]byte, error) {
	fg := color.RGBA{R: c.textColor.R, G: c.textColor.G, B: c.textColor.B, A: 255}
	wide := font.RenderStringImage(c.font, c.text, c.displaySize, fg)

	w := c.displaySize.Width
	h := c.displaySize.Height
	textWidth := wide.Bounds().Dx()

	// Total scroll distance: start offscreen right, scroll until offscreen left
	totalScroll := w + textWidth

	var frames []*image.Paletted
	var delays []int

	pal := buildPalette(fg)

	for offset := 0; offset < totalScroll; offset += c.pixelsPerFrame {
		frame := image.NewPaletted(image.Rect(0, 0, w, h), pal)
		// Fill black
		draw.Draw(frame, frame.Bounds(), image.Black, image.Point{}, draw.Src)

		// The text starts at x = w (offscreen right) and moves left
		srcX := offset - w
		if srcX < 0 {
			// Text hasn't fully entered yet; draw partial
			dstX := -srcX
			sr := image.Rect(0, 0, w-dstX, h)
			draw.Draw(frame, image.Rect(dstX, 0, w, h), wide, sr.Min, draw.Over)
		} else if srcX+w <= textWidth {
			// Text fully visible window
			sr := image.Rect(srcX, 0, srcX+w, h)
			draw.Draw(frame, frame.Bounds(), wide, sr.Min, draw.Over)
		} else {
			// Text exiting left side
			remaining := textWidth - srcX
			if remaining > 0 {
				sr := image.Rect(srcX, 0, textWidth, h)
				draw.Draw(frame, image.Rect(0, 0, remaining, h), wide, sr.Min, draw.Over)
			}
		}

		frames = append(frames, frame)
		delays = append(delays, c.frameDelay)
	}

	var buf bytes.Buffer
	err := gif.EncodeAll(&buf, &gif.GIF{
		Image: frames,
		Delay: delays,
		Config: image.Config{
			Width:      w,
			Height:     h,
			ColorModel: pal,
		},
		LoopCount: 0, // loop forever
	})
	if err != nil {
		return nil, err
	}

	return protocol.EncodeGIF(buf.Bytes(), c.displaySize)
}

// buildPalette creates a minimal palette: black, text color, and two
// intermediate shades for anti-aliasing. Fewer colors = much smaller GIF.
func buildPalette(fg color.RGBA) color.Palette {
	return color.Palette{
		color.RGBA{0, 0, 0, 255},
		color.RGBA{fg.R / 3, fg.G / 3, fg.B / 3, 255},
		color.RGBA{fg.R * 2 / 3, fg.G * 2 / 3, fg.B * 2 / 3, 255},
		fg,
	}
}

func WithScrollTextColor(c protocol.Color) ScrollTextOption {
	return func(cmd *scrollTextCmd) { cmd.textColor = c }
}

func WithScrollFont(r font.Renderer) ScrollTextOption {
	return func(cmd *scrollTextCmd) { cmd.font = r }
}

func WithScrollDisplaySize(s protocol.DisplaySize) ScrollTextOption {
	return func(cmd *scrollTextCmd) { cmd.displaySize = s }
}

func WithScrollSpeed(pixelsPerFrame, delayMs int) ScrollTextOption {
	return func(cmd *scrollTextCmd) {
		if pixelsPerFrame > 0 {
			cmd.pixelsPerFrame = pixelsPerFrame
		}
		if delayMs > 0 {
			cmd.frameDelay = delayMs / 10 // convert ms to gif 1/100s units
			if cmd.frameDelay < 1 {
				cmd.frameDelay = 1
			}
		}
	}
}
