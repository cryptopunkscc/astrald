package channel

import (
	"io"

	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/log"
)

type RenderSender struct {
	w io.Writer
	p *log.Printer
}

var _ Sender = &RenderSender{}

func NewRenderSender(w io.Writer) *RenderSender {
	return &RenderSender{
		w: w,
		p: log.NewPrinter(w),
	}
}

func (p RenderSender) Send(object astral.Object) error {
	_, err := p.w.Write([]byte(log.DefaultViewer.Render(object) + "\n"))
	return err
}
