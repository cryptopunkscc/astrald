package color

import (
	"fmt"
	"hash/fnv"
	"math/rand"
)

// TextColorFromString returns a deterministic 6-digit lowercase hex color
// (e.g. "d070ea") based on the input string.
// The color is vibrant and has good readability on both dark and light backgrounds.
func TextColorFromString(s string) string {
	if s == "" {
		s = "default"
	}

	// Create a deterministic seed from the string
	h := fnv.New64a()
	h.Write([]byte(s))
	seed := h.Sum64()

	// Use a new Rand instance with the seed
	r := rand.New(rand.NewSource(int64(seed)))

	// Hue: full spectrum
	hue := r.Intn(360)

	// High saturation for vivid colors (never grey)
	saturation := 48 + r.Intn(48) // 48–95%

	// Lightness in the sweet spot for text contrast on both backgrounds
	lightness := 40 + r.Intn(36) // 40–75%

	// Convert HSL → RGB
	rgbR, rgbG, rgbB := hslToRGB(float64(hue), float64(saturation)/100.0, float64(lightness)/100.0)

	return fmt.Sprintf("#%02x%02x%02x", rgbR, rgbG, rgbB)
}

// hslToRGB converts HSL (0-360, 0-1, 0-1) to RGB (0-255)
func hslToRGB(h, s, l float64) (r, g, b int) {
	if s == 0 {
		v := int(l * 255)
		return v, v, v
	}

	var q float64
	if l < 0.5 {
		q = l * (1 + s)
	} else {
		q = l + s - l*s
	}
	p := 2*l - q

	r = int(255 * hueToRGB(p, q, h/360+1.0/3))
	g = int(255 * hueToRGB(p, q, h/360))
	b = int(255 * hueToRGB(p, q, h/360-1.0/3))

	return
}

func hueToRGB(p, q, t float64) float64 {
	if t < 0 {
		t += 1
	}
	if t > 1 {
		t -= 1
	}
	if t < 1.0/6 {
		return p + (q-p)*6*t
	}
	if t < 0.5 {
		return q
	}
	if t < 2.0/3 {
		return p + (q-p)*(2.0/3-t)*6
	}
	return p
}
