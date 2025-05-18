package objects

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/mod/objects"
)

type WriterWrapper struct {
	mod *Module
	objects.Writer
}

func NewWriterWrapper(mod *Module, w objects.Writer) *WriterWrapper {
	return &WriterWrapper{mod: mod, Writer: w}
}

func (w *WriterWrapper) Write(p []byte) (n int, err error) {
	return w.Writer.Write(p)

}

func (w *WriterWrapper) Discard() error {
	return w.Writer.Discard()
}

func (w *WriterWrapper) Commit() (*astral.ObjectID, error) {
	objectID, err := w.Writer.Commit()
	if err != nil {
		w.mod.Receive(&objects.EventCommitted{ObjectID: objectID}, nil)
	}

	return objectID, err
}
