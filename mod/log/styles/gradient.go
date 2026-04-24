package styles

import "strings"

type Gradient struct {
	From Color
	To   Color
}

func NewGradient(from Color, to Color) Gradient {
	return Gradient{From: from, To: to}
}

func (g Gradient) Bri(factor float64) Gradient {
	return Gradient{
		From: g.From.Bri(factor),
		To:   g.To.Bri(factor),
	}
}

func (g Gradient) Sat(factor float64) Gradient {
	return Gradient{
		From: g.From.Sat(factor),
		To:   g.To.Sat(factor),
	}
}

func (g Gradient) Hue(offset float64) Gradient {
	return Gradient{
		From: g.From.Hue(offset),
		To:   g.To.Hue(offset),
	}
}

func (g Gradient) Render(s ...string) string {
	var result []string
	for _, str := range s {
		if len(str) == 0 {
			result = append(result, str)
			continue
		}

		var colored string
		runes := []rune(str)
		length := len(runes)

		for i, r := range runes {
			var ratio float64
			if length == 1 {
				ratio = 0
			} else {
				ratio = float64(i) / float64(length-1)
			}

			color := interpolate(g.From, g.To, ratio)
			colored += color.Render(string(r))
		}
		result = append(result, colored)
	}

	return strings.Join(result, " ")
}

func interpolate(from Color, to Color, ratio float64) Color {
	return Color{
		H: from.Hue((to.H - from.H) * ratio).H,
		S: from.S + (to.S-from.S)*ratio,
		L: from.L + (to.L-from.L)*ratio,
		A: from.A + (to.A-from.A)*ratio,
	}
}
