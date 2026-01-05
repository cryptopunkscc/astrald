package log

import (
	"encoding"
	"fmt"

	"github.com/cryptopunkscc/astrald/astral"
)

type Renderer interface {
	Render() string
}

func Render(objects ...astral.Object) (s string) {
	for _, o := range objects {
		if o == nil {
			o = &astral.Nil{}
		}

		if r, ok := o.(Renderer); ok {
			s += r.Render()
			continue
		}

		if r, ok := o.(encoding.TextMarshaler); ok {
			var text []byte
			text, _ = r.MarshalText()
			s += string(text)
			continue
		}

		if r, ok := o.(fmt.Stringer); ok {
			s += r.String()
			continue
		}

		s += fmt.Sprintf("[[%s]]", o.ObjectType())
	}

	return
}
