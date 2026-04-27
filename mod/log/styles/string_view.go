package styles

import (
	"io"

	"github.com/cryptopunkscc/astrald/astral"
)

type StringView struct {
	Style Renderer
	str   *astral.String32
}

var _ astral.Object = &StringView{}

// astral:blueprint-ignore
func (StringView) ObjectType() string {
	return astral.String32("").ObjectType()
}

func (v StringView) Render() string {
	return v.Style.Render(v.str.String())
}

func (v StringView) WriteTo(writer io.Writer) (n int64, err error) {
	return v.str.WriteTo(writer)
}

func (v *StringView) ReadFrom(reader io.Reader) (n int64, err error) {
	return v.str.ReadFrom(reader)
}

func (v StringView) String() string {
	return v.str.String()
}

func String(text string, style Renderer) *StringView {
	return &StringView{
		Style: style,
		str:   astral.NewString32(text),
	}
}
