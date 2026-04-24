package styles

import "github.com/charmbracelet/lipgloss"

type Style struct {
	text Color
}

func NewStyle() Style {
	return Style{text: White}
}

func (s Style) Text(c Color) Style {
	s.text = c
	return s
}

func (s Style) Render(text string) string {
	l := lipgloss.NewStyle()
	l = l.Foreground(lipgloss.Color(s.text.RGB()))
	return l.Render(text)
}
