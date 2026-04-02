package protocol

// DisplaySize represents the pixel dimensions of an iDotMatrix display.
type DisplaySize struct {
	Width  int
	Height int
}

var (
	Display16x16 = DisplaySize{16, 16}
	Display32x32 = DisplaySize{32, 32}
	Display64x64 = DisplaySize{64, 64}
)

// TextMode controls how text is animated on the display.
type TextMode uint8

const (
	TextFixed       TextMode = 0
	TextScrollLeft  TextMode = 1
	TextScrollRight TextMode = 2
	TextScrollUp    TextMode = 3
	TextScrollDown  TextMode = 4
	TextStrobe      TextMode = 5
	TextFade        TextMode = 6
	TextFalling     TextMode = 7
	TextLaser       TextMode = 8
)

// ColorMode controls how text color is applied.
type ColorMode uint8

const (
	ColorFixed      ColorMode = 1
	ColorBlueRed    ColorMode = 2
	ColorPastel     ColorMode = 3
	ColorPinkOrange ColorMode = 4
)

// Color represents an RGB color value.
type Color struct {
	R, G, B uint8
}

// TextOptions configures text display parameters.
type TextOptions struct {
	Mode       TextMode
	Speed      uint8
	ColorMode  ColorMode
	TextColor  Color
	Background *Color // nil = no background
}
