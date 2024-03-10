package storage

import (
	"errors"
	"github.com/cryptopunkscc/astrald/cslq"
	"github.com/cryptopunkscc/astrald/data"
	"github.com/cryptopunkscc/astrald/mod/storage"
	api "github.com/cryptopunkscc/astrald/mod/storage/cslq"
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
	case api.WriterDiscard:
		if err = w.writer.Discard(); err != nil {
			log.Println(err)
		}
		return ErrClose
	case api.WriterWrite:
		var b []byte
		if err = cslq.Decode(w.conn, "[l]c", &b); err != nil {
			return
		}
		var l int
		if l, err = w.writer.Write(b); err != nil {
			return
		}
		if err = cslq.Encode(w.conn, "l", l); err != nil {
			return
		}
	case api.WriterCommit:
		var dataID data.ID
		dataID, err = w.writer.Commit()
		if err = cslq.Encode(w.conn, "v", dataID); err != nil {
			return
		}
		return ErrClose
	}
	return
}
