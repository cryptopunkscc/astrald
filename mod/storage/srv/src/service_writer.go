package srv

import (
	"encoding/binary"
	"errors"
	"github.com/cryptopunkscc/astrald/data"
	"github.com/cryptopunkscc/astrald/mod/storage"
	proto "github.com/cryptopunkscc/astrald/mod/storage/srv"
	"io"
	"log"
)

type WriterService struct {
	writer storage.Writer
	conn   proto.BinaryEncoder
}

func NewWriterService(writer storage.Writer, conn io.ReadWriter) *WriterService {
	return &WriterService{writer, proto.NewBinaryEncoder(conn)}
}

func (w *WriterService) Loop() (err error) {
	for {
		var cmd byte
		if err = binary.Read(w.conn, binary.BigEndian, &cmd); err != nil {
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

func (w *WriterService) Handle(cmd byte) (err error) {
	switch cmd {
	case proto.WriterWrite:
		var n32 int32
		if err = w.conn.Decode(&n32); err != nil {
			return
		}
		b := make([]byte, n32)
		if err = w.conn.Decode(&b); err != nil {
			return
		}
		var n int
		if n, err = w.writer.Write(b); err != nil {
			return
		}
		if err = w.conn.Encode(int32(n)); err != nil {
			return
		}
	case proto.WriterDiscard:
		if err = w.writer.Discard(); err != nil {
			log.Println(err)
		}
		return ErrClose
	case proto.WriterCommit:
		var dataID data.ID
		dataID, err = w.writer.Commit()
		if err = w.conn.Encode(dataID); err != nil {
			return
		}
		return ErrClose
	}
	return
}
