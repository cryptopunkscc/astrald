package log

import (
	"github.com/charmbracelet/lipgloss"
	"github.com/cryptopunkscc/astrald/astral"
)

type StringView struct {
	Style *lipgloss.Style
	*astral.String
}

func (StringView) ObjectType() string { return "" }

func (v StringView) Render() string {
	return v.Style.Render(v.String.String())
}

func String(text string, style *lipgloss.Style) *StringView {
	return &StringView{
		Style:  style,
		String: astral.NewString(text),
	}
}
