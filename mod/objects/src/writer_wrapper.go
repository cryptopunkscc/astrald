package objects

import (
	"github.com/cryptopunkscc/astrald/mod/objects"
	"github.com/cryptopunkscc/astrald/object"
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

func (w *WriterWrapper) Commit() (object.ID, error) {
	objectID, err := w.Writer.Commit()
	if err != nil {
		w.mod.events.Emit(objects.EventObjectCommitted{ObjectID: objectID})
	}

	return objectID, err
}
