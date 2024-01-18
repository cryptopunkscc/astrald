package storage

import (
	"github.com/cryptopunkscc/astrald/data"
	"github.com/cryptopunkscc/astrald/mod/storage"
)

type DataWriterWrapper struct {
	mod *Module
	storage.DataWriter
}

func NewDataWriterWrapper(mod *Module, dataWriter storage.DataWriter) *DataWriterWrapper {
	return &DataWriterWrapper{mod: mod, DataWriter: dataWriter}
}

func (w *DataWriterWrapper) Write(p []byte) (n int, err error) {
	return w.DataWriter.Write(p)

}

func (w *DataWriterWrapper) Discard() error {
	return w.DataWriter.Discard()
}

func (w *DataWriterWrapper) Commit() (data.ID, error) {
	dataID, err := w.DataWriter.Commit()
	if err != nil {
		w.mod.events.Emit(storage.EventDataCommitted{DataID: dataID})
	}

	return dataID, err
}
