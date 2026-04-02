package main

import (
	"fmt"
	"math"
	"os"

	"github.com/jeffory/idotmatrix-lib/command"
	"github.com/jeffory/idotmatrix-lib/device"
	"github.com/jeffory/idotmatrix-lib/protocol"
	"github.com/jeffory/idotmatrix-lib/transport"
)

func main() {
	addr := os.Getenv("IDOTMATRIX_ADDRESS")
	if addr == "" {
		fmt.Println("Set IDOTMATRIX_ADDRESS env var to device MAC")
		return
	}
	conn, err := transport.Connect(addr, transport.DefaultOptions())
	if err != nil {
		fmt.Printf("Connect error: %v\n", err)
		return
	}
	defer conn.Close()

	dev := device.New(conn, protocol.Display32x32)

	// Draw a rainbow spiral
	cx, cy := 16, 16
	colorIdx := 0
	for t := range 500 {
		angle := 0.1 * float64(t)
		radius := 0.5 * angle
		x := int(float64(cx) + radius*math.Cos(angle))
		y := int(float64(cy) + radius*math.Sin(angle))
		if x < 0 || x >= 32 || y < 0 || y >= 32 {
			continue
		}
		h := float64(colorIdx%31) * 360.0 / 31.0
		r, g, b := hsvToRGB(h, 1.0, 1.0)
		cmd := command.DrawPixel(uint8(x), uint8(y), protocol.Color{R: r, G: g, B: b})
		if err := dev.Send(cmd); err != nil {
			fmt.Printf("Error: %v\n", err)
			return
		}
		colorIdx++
	}
	fmt.Printf("Drew %d pixels\n", colorIdx)
}

func hsvToRGB(h, s, v float64) (uint8, uint8, uint8) {
	c := v * s
	x := c * (1.0 - math.Abs(math.Mod(h/60.0, 2.0)-1.0))
	m := v - c
	var r, g, b float64
	switch {
	case h < 60:
		r, g, b = c, x, 0
	case h < 120:
		r, g, b = x, c, 0
	case h < 180:
		r, g, b = 0, c, x
	case h < 240:
		r, g, b = 0, x, c
	case h < 300:
		r, g, b = x, 0, c
	default:
		r, g, b = c, 0, x
	}
	return uint8((r + m) * 255), uint8((g + m) * 255), uint8((b + m) * 255)
}
