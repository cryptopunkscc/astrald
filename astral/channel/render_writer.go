package channel

import (
	"io"

	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/log"
)

type RenderWriter struct {
	w io.Writer
	p *log.Printer
}

var _ Writer = &RenderWriter{}

func NewRenderWriter(w io.Writer) *RenderWriter {
	return &RenderWriter{
		w: w,
		p: log.NewPrinter(w),
	}
}

func (p RenderWriter) Write(object astral.Object) error {
	_, err := p.w.Write([]byte(log.DefaultViewer.Render(object) + "\n"))
	return err
}
