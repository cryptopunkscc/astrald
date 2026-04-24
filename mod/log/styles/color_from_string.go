package styles

import (
	"hash/fnv"
	"math/rand"
)

// ColorFromString returns a deterministic 6-digit lowercase hex color
// (e.g. "d070ea") based on the input string.
// The color is vibrant and has good readability on both dark and light backgrounds.
func ColorFromString(s string) Color {
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

	return HSL(float64(hue), float64(saturation), float64(lightness))
}
