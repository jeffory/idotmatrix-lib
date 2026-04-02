package command

import "image"

// AdjustImage applies contrast and saturation adjustments to an RGBA image
// in-place. Both values are percentages from -100 to 100, where 0 means no
// change. Positive values increase contrast/saturation, negative values decrease.
func AdjustImage(img *image.RGBA, contrast, saturation float64) {
	if contrast == 0 && saturation == 0 {
		return
	}

	contrastFactor := 1 + contrast/100
	satFactor := 1 + saturation/100
	pix := img.Pix

	for i := 0; i < len(pix); i += 4 {
		r := float64(pix[i])
		g := float64(pix[i+1])
		b := float64(pix[i+2])

		// Contrast: scale distance from midpoint (128)
		if contrast != 0 {
			r = 128 + contrastFactor*(r-128)
			g = 128 + contrastFactor*(g-128)
			b = 128 + contrastFactor*(b-128)
		}

		// Saturation: scale distance from luminance
		if saturation != 0 {
			gray := 0.299*r + 0.587*g + 0.114*b
			r = gray + satFactor*(r-gray)
			g = gray + satFactor*(g-gray)
			b = gray + satFactor*(b-gray)
		}

		pix[i] = clampByte(r)
		pix[i+1] = clampByte(g)
		pix[i+2] = clampByte(b)
	}
}

func clampByte(v float64) uint8 {
	if v < 0 {
		return 0
	}
	if v > 255 {
		return 255
	}
	return uint8(v)
}
