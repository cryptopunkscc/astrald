package storage

import (
	"github.com/cryptopunkscc/astrald/data"
	"github.com/cryptopunkscc/astrald/mod/storage"
)

type DataWriterWrapper struct {
	mod *Module
	storage.Writer
}

func NewDataWriterWrapper(mod *Module, dataWriter storage.Writer) *DataWriterWrapper {
	return &DataWriterWrapper{mod: mod, Writer: dataWriter}
}

func (w *DataWriterWrapper) Write(p []byte) (n int, err error) {
	return w.Writer.Write(p)

}

func (w *DataWriterWrapper) Discard() error {
	return w.Writer.Discard()
}

func (w *DataWriterWrapper) Commit() (data.ID, error) {
	dataID, err := w.Writer.Commit()
	if err != nil {
		w.mod.events.Emit(storage.EventDataCommitted{DataID: dataID})
	}

	return dataID, err
}
