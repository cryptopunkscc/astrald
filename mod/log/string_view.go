package log

import (
	"errors"
	"io"

	"github.com/charmbracelet/lipgloss"
	"github.com/cryptopunkscc/astrald/astral"
)

type StringView struct {
	Style  *lipgloss.Style
	String *astral.String32
}

var _ astral.Object = &StringView{}

// astral:blueprint-ignore
func (StringView) ObjectType() string { return "" }

func (v StringView) Render() string {
	return v.Style.Render(v.String.String())
}

func (StringView) WriteTo(writer io.Writer) (n int64, err error) {
	return 0, errors.New("WriteTo called on a pseudo-object")
}

func (StringView) ReadFrom(reader io.Reader) (n int64, err error) {
	return 0, errors.New("WriteTo called on a pseudo-object")
}

func String(text string, style *lipgloss.Style) *StringView {
	return &StringView{
		Style:  style,
		String: astral.NewString32(text),
	}
}
