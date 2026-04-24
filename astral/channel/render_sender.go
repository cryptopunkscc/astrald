package channel

import (
	"io"

	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/fmt"
)

type RenderSender struct {
	w io.Writer
}

var _ Sender = &RenderSender{}

func NewRenderSender(w io.Writer) *RenderSender {
	return &RenderSender{
		w: w,
	}
}

func (p RenderSender) Send(object astral.Object) error {
	_, err := p.w.Write([]byte(fmt.Sprint(object) + "\n"))
	return err
}
