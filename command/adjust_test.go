package command

import (
	"image"
	"image/color"
	"testing"
)

func solidImage(c color.RGBA) *image.RGBA {
	img := image.NewRGBA(image.Rect(0, 0, 2, 2))
	for i := 0; i < len(img.Pix); i += 4 {
		img.Pix[i] = c.R
		img.Pix[i+1] = c.G
		img.Pix[i+2] = c.B
		img.Pix[i+3] = c.A
	}
	return img
}

func TestAdjustImage_NoChange(t *testing.T) {
	img := solidImage(color.RGBA{R: 100, G: 150, B: 200, A: 255})
	AdjustImage(img, 0, 0)
	if img.Pix[0] != 100 || img.Pix[1] != 150 || img.Pix[2] != 200 {
		t.Errorf("expected no change, got R=%d G=%d B=%d", img.Pix[0], img.Pix[1], img.Pix[2])
	}
}

func TestAdjustImage_PositiveContrast(t *testing.T) {
	// Red=200 is above midpoint (128), should increase
	img := solidImage(color.RGBA{R: 200, G: 100, B: 128, A: 255})
	AdjustImage(img, 50, 0)

	if img.Pix[0] <= 200 {
		t.Errorf("expected R > 200 with positive contrast, got %d", img.Pix[0])
	}
	if img.Pix[1] >= 100 {
		t.Errorf("expected G < 100 with positive contrast (below midpoint), got %d", img.Pix[1])
	}
	// Blue=128 is at midpoint, should stay at 128
	if img.Pix[2] != 128 {
		t.Errorf("expected B=128 (at midpoint), got %d", img.Pix[2])
	}
}

func TestAdjustImage_PositiveSaturation(t *testing.T) {
	// A red pixel should become more red (further from gray)
	img := solidImage(color.RGBA{R: 200, G: 50, B: 50, A: 255})
	AdjustImage(img, 0, 50)

	if img.Pix[0] <= 200 {
		t.Errorf("expected R to increase with saturation boost, got %d", img.Pix[0])
	}
	if img.Pix[1] >= 50 {
		t.Errorf("expected G to decrease with saturation boost, got %d", img.Pix[1])
	}
}

func TestAdjustImage_NegativeSaturation(t *testing.T) {
	// Full desaturation (-100) should produce gray
	img := solidImage(color.RGBA{R: 200, G: 50, B: 50, A: 255})
	AdjustImage(img, 0, -100)

	// With -100 saturation, factor = 0, all channels should equal luminance
	// gray = 0.299*200 + 0.587*50 + 0.114*50 = 59.8 + 29.35 + 5.7 = 94.85 ≈ 95
	r, g, b := img.Pix[0], img.Pix[1], img.Pix[2]
	if r != g || g != b {
		t.Errorf("expected grayscale with -100 saturation, got R=%d G=%d B=%d", r, g, b)
	}
}

func TestAdjustImage_ClampsAt255(t *testing.T) {
	img := solidImage(color.RGBA{R: 250, G: 250, B: 250, A: 255})
	AdjustImage(img, 100, 0) // double contrast on near-white

	if img.Pix[0] != 255 {
		t.Errorf("expected R clamped at 255, got %d", img.Pix[0])
	}
}

func TestAdjustImage_ClampsAt0(t *testing.T) {
	img := solidImage(color.RGBA{R: 5, G: 5, B: 5, A: 255})
	AdjustImage(img, 100, 0) // double contrast on near-black

	if img.Pix[0] != 0 {
		t.Errorf("expected R clamped at 0, got %d", img.Pix[0])
	}
}

func TestAdjustImage_PreservesAlpha(t *testing.T) {
	img := solidImage(color.RGBA{R: 100, G: 100, B: 100, A: 128})
	AdjustImage(img, 50, 50)

	if img.Pix[3] != 128 {
		t.Errorf("expected alpha preserved at 128, got %d", img.Pix[3])
	}
}
