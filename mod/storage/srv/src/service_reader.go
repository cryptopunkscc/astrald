package srv

import (
	"encoding/binary"
	"encoding/json"
	"errors"
	"github.com/cryptopunkscc/astrald/mod/storage"
	proto "github.com/cryptopunkscc/astrald/mod/storage/srv"
	"io"
	"log"
)

var ErrClose = errors.New("close")

type ReaderService struct {
	reader storage.Reader
	conn   proto.BinaryEncoder
}

func NewReaderService(reader storage.Reader, conn io.ReadWriter) *ReaderService {
	return &ReaderService{reader: reader, conn: proto.NewBinaryEncoder(conn)}
}

func (srv *ReaderService) Loop() (err error) {
	for {
		var cmd byte
		if err = binary.Read(srv.conn, binary.BigEndian, &cmd); err != nil {
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

func (srv *ReaderService) Handle(cmd byte) (err error) {
	switch cmd {
	case proto.ReaderRead:
		var n32 int32
		if err = srv.conn.Decode(&n32); err != nil {
			return
		}
		if _, err = io.CopyN(srv.conn, srv.reader, int64(n32)); err != nil {
			return
		}
	case proto.ReaderClose:
		if err = srv.reader.Close(); err != nil {
			log.Println(err)
		}
		return ErrClose
	case proto.ReaderSeek:
		var offset int64
		var whence int8
		var seek int64
		if err = srv.conn.Decode(&offset, &whence); err != nil {
			return
		}
		if seek, err = srv.reader.Seek(offset, int(whence)); err != nil {
			return
		}
		if err = srv.conn.Encode(seek); err != nil {
			return
		}
	case proto.ReaderInfo:
		info := srv.reader.Info()
		if err = json.NewEncoder(srv.conn).Encode(info); err != nil {
			return
		}
	}
	return
}
