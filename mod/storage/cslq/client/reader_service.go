package storage

import (
	"errors"
	"github.com/cryptopunkscc/astrald/cslq"
	"github.com/cryptopunkscc/astrald/mod/storage"
	"io"
	"log"
)

var ErrClose = errors.New("close")

type readerService struct {
	reader storage.Reader
	conn   io.ReadWriter
}

func newReaderService(reader storage.Reader, conn io.ReadWriter) *readerService {
	return &readerService{reader: reader, conn: conn}
}

func (srv *readerService) Loop() (err error) {
	for {
		var cmd byte
		if err = cslq.Decode(srv.conn, "c", &cmd); err != nil {
			return
		}
		if err = srv.Handle(cmd); err != nil {
			if errors.Is(err, ErrClose) {
				err = nil
			}
			return
		}
	}
}

func (srv *readerService) Handle(cmd byte) (err error) {
	switch cmd {
	case ReaderClose:
		if err = srv.reader.Close(); err != nil {
			log.Println(err)
		}
		return ErrClose
	case ReaderRead:
		var l int
		if err = cslq.Decode(srv.conn, "l", &l); err != nil {
			return
		}
		if _, err = io.CopyN(srv.conn, srv.reader, int64(l)); err != nil {
			return
		}
	case ReaderInfo:
		name := srv.reader.Info().Name
		if err = cslq.Encode(srv.conn, "[c]c", name); err != nil {
			return
		}
	case ReaderSeek:
		var offset int64
		var whence int
		var seek int64
		if err = cslq.Decode(srv.conn, "q l", &offset, &whence); err != nil {
			return
		}
		if seek, err = srv.reader.Seek(offset, whence); err != nil {
			return
		}
		if err = cslq.Encode(srv.conn, "q", seek); err != nil {
			return
		}
	}
	return
}
