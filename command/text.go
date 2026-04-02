package command

import (
	"github.com/jeffory/idotmatrix-lib/font"
	"github.com/jeffory/idotmatrix-lib/protocol"
)

type TextOption func(*textCmd)

type textCmd struct {
	text        string
	mode        protocol.TextMode
	speed       uint8
	colorMode   protocol.ColorMode
	textColor   protocol.Color
	background  *protocol.Color
	font        font.Renderer
	displaySize protocol.DisplaySize
}

func NewText(text string, opts ...TextOption) Command {
	cmd := &textCmd{
		text:        text,
		mode:        protocol.TextFixed,
		speed:       100,
		colorMode:   protocol.ColorFixed,
		textColor:   protocol.Color{R: 255, G: 255, B: 255},
		displaySize: protocol.Display32x32,
		font:        font.DefaultBitmap(),
	}
	for _, opt := range opts {
		opt(cmd)
	}
	return cmd
}

func (*textCmd) Chunked() bool { return true }

func (c *textCmd) Encode() ([][]byte, error) {
	bitmaps, err := c.font.RenderString(c.text, c.displaySize)
	if err != nil {
		return nil, err
	}
	opts := protocol.TextOptions{
		Mode:       c.mode,
		Speed:      c.speed,
		ColorMode:  c.colorMode,
		TextColor:  c.textColor,
		Background: c.background,
	}
	return protocol.EncodeText(bitmaps, opts)
}

func WithTextMode(m protocol.TextMode) TextOption      { return func(c *textCmd) { c.mode = m } }
func WithSpeed(s uint8) TextOption                     { return func(c *textCmd) { c.speed = s } }
func WithColorMode(m protocol.ColorMode) TextOption    { return func(c *textCmd) { c.colorMode = m } }
func WithTextColor(color protocol.Color) TextOption    { return func(c *textCmd) { c.textColor = color } }
func WithBackground(color protocol.Color) TextOption   { return func(c *textCmd) { c.background = &color } }
func WithFont(r font.Renderer) TextOption              { return func(c *textCmd) { c.font = r } }
func WithDisplaySize(s protocol.DisplaySize) TextOption { return func(c *textCmd) { c.displaySize = s } }
