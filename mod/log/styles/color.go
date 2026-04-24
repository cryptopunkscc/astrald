package styles

import (
	"fmt"
	"image/color"
	"math"
	"strconv"

	"github.com/charmbracelet/lipgloss"
)

type Color struct {
	H, S, L, A float64
}

var (
	Black     = HSL(0, 0, 0)
	DarkGray  = HSL(0, 0, 30)
	Gray      = HSL(0, 0, 50)
	White     = HSL(0, 0, 70)
	Highlight = HSL(0, 0, 100)

	Red     = HSL(0, 100, 40)
	Orange  = HSL(30, 100, 40)
	Yellow  = HSL(60, 100, 40)
	Lime    = HSL(90, 100, 40)
	Green   = HSL(120, 100, 40)
	Mint    = HSL(150, 100, 40)
	Cyan    = HSL(180, 100, 40)
	Azure   = HSL(210, 100, 40)
	Blue    = HSL(240, 100, 40)
	Violet  = HSL(270, 100, 45)
	Magenta = HSL(300, 100, 40)
	Rose    = HSL(330, 100, 40)
)

func RGBA(hex string) Color {
	rgba := parseRGBA(hex)

	// Normalize RGB values to 0-1 range
	r := float64(rgba.R) / 255.0
	g := float64(rgba.G) / 255.0
	b := float64(rgba.B) / 255.0
	a := float64(rgba.A) / 255.0

	// Find min and max values
	maxVal := math.Max(math.Max(r, g), b)
	minVal := math.Min(math.Min(r, g), b)
	delta := maxVal - minVal

	// Calculate lightness
	l := (maxVal + minVal) / 2.0

	// Calculate saturation
	var s float64
	if delta == 0 {
		s = 0
	} else {
		if l < 0.5 {
			s = delta / (maxVal + minVal)
		} else {
			s = delta / (2.0 - maxVal - minVal)
		}
	}

	// Calculate hue
	var h float64
	if delta == 0 {
		h = 0
	} else {
		switch maxVal {
		case r:
			h = ((g - b) / delta)
			if g < b {
				h += 6.0
			}
		case g:
			h = ((b - r) / delta) + 2.0
		case b:
			h = ((r - g) / delta) + 4.0
		}
		h /= 6.0
	}

	return Color{H: h, S: s, L: l, A: a}
}

func HSL(h, s, l float64) Color {
	return Color{H: h / 360, S: s / 100, L: l / 100}
}

func (c Color) RGB() string {
	h := c.H
	s := c.S
	l := c.L

	var r, g, b float64

	if s == 0 {
		// Achromatic (gray)
		r = l
		g = l
		b = l
	} else {
		hue2rgb := func(p, q, t float64) float64 {
			if t < 0 {
				t += 1
			}
			if t > 1 {
				t -= 1
			}
			if t < 1.0/6.0 {
				return p + (q-p)*6*t
			}
			if t < 1.0/2.0 {
				return q
			}
			if t < 2.0/3.0 {
				return p + (q-p)*(2.0/3.0-t)*6
			}
			return p
		}

		var q float64
		if l < 0.5 {
			q = l * (1 + s)
		} else {
			q = l + s - l*s
		}
		p := 2*l - q

		r = hue2rgb(p, q, h+1.0/3.0)
		g = hue2rgb(p, q, h)
		b = hue2rgb(p, q, h-1.0/3.0)
	}

	// Convert to 0-255 range
	r8 := uint8(math.Round(r * 255))
	g8 := uint8(math.Round(g * 255))
	b8 := uint8(math.Round(b * 255))

	return fmt.Sprintf("#%02x%02x%02x", r8, g8, b8)
}

func (c Color) Bri(factor float64) Color {
	c.L = max(min(c.L*factor, 1.0), 0)
	return c
}

func (c Color) Sat(factor float64) Color {
	c.S = max(min(c.S*factor, 1.0), 0)
	return c
}

func (c Color) Hue(offset float64) Color {
	c.H += offset
	if c.H < 0 {
		c.H -= float64(int(c.H) + 1)
	}
	for c.H > 1 {
		c.H -= float64(int(c.H))
	}
	return c
}

func (c Color) Complement() Color {
	return c.Hue(0.5)
}

func (c Color) Triad() (Color, Color) {
	return c.Hue(1.0 / 3.0), c.Hue(2.0 / 3.0)
}

func (c Color) Tetrad() (Color, Color, Color) {
	return c.Hue(1.0 / 4.0), c.Hue(2.0 / 4.0), c.Hue(3.0 / 4.0)
}

func (c Color) Render(text ...string) string {
	l := lipgloss.NewStyle()
	l = l.Foreground(lipgloss.Color(c.RGB()))
	return l.Render(text...)
}

func parseRGBA(s string) color.RGBA {
	// Remove leading '#' if present
	if len(s) > 0 && s[0] == '#' {
		s = s[1:]
	}

	var r, g, b, a uint8 = 0, 0, 0, 255

	if len(s) == 6 {
		// RRGGBB format
		if val, err := strconv.ParseUint(s[0:2], 16, 8); err == nil {
			r = uint8(val)
		}
		if val, err := strconv.ParseUint(s[2:4], 16, 8); err == nil {
			g = uint8(val)
		}
		if val, err := strconv.ParseUint(s[4:6], 16, 8); err == nil {
			b = uint8(val)
		}
	} else if len(s) == 8 {
		// RRGGBBAA format
		if val, err := strconv.ParseUint(s[0:2], 16, 8); err == nil {
			r = uint8(val)
		}
		if val, err := strconv.ParseUint(s[2:4], 16, 8); err == nil {
			g = uint8(val)
		}
		if val, err := strconv.ParseUint(s[4:6], 16, 8); err == nil {
			b = uint8(val)
		}
		if val, err := strconv.ParseUint(s[6:8], 16, 8); err == nil {
			a = uint8(val)
		}
	}

	return color.RGBA{R: r, G: g, B: b, A: a}
}
