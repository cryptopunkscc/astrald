package storage

import (
	"errors"
	"github.com/cryptopunkscc/astrald/cslq"
	"github.com/cryptopunkscc/astrald/data"
	"github.com/cryptopunkscc/astrald/mod/storage"
	"io"
	"log"
)

type writerService struct {
	writer storage.Writer
	conn   io.ReadWriter
}

func newWriterService(writer storage.Writer, conn io.ReadWriter) *writerService {
	return &writerService{writer: writer, conn: conn}
}

func (w *writerService) Loop() (err error) {
	for {
		var cmd byte
		if err = cslq.Decode(w.conn, "c", &cmd); err != nil {
			return
		}
		if err = w.Handle(cmd); err != nil {
			if errors.Is(err, ErrClose) {
				err = nil
			}
			return
		}
	}
}

func (w *writerService) Handle(cmd byte) (err error) {
	switch cmd {
	case WriterDiscard:
		if err = w.writer.Discard(); err != nil {
			log.Println(err)
		}
		return ErrClose
	case WriterWrite:
		var l int
		if err = cslq.Decode(w.conn, "c", &l); err != nil {
			return
		}
		if _, err = io.CopyN(w.writer, w.conn, int64(l)); err != nil {
			return
		}
	case WriterCommit:
		var dataID data.ID
		dataID, err = w.writer.Commit()
		if err = cslq.Encode(w.conn, "v", dataID); err != nil {
			return
		}
	}
	return
}
