package storage

import (
	"encoding/json"
	"github.com/cryptopunkscc/astrald/mod/storage"
	proto "github.com/cryptopunkscc/astrald/mod/storage/srv"
	"io"
)

type ReaderClient struct {
	conn proto.BinaryEncoderCloser
}

func NewReaderClient(conn io.ReadWriteCloser) storage.Reader {
	return &ReaderClient{proto.NewBinaryEncoderCloser(conn)}
}

func (r *ReaderClient) Read(p []byte) (n int, err error) {
	if err = r.conn.Encode(proto.ReaderRead, int32(cap(p))); err != nil {
		return
	}
	n, err = r.conn.Read(p)
	return
}

func (r *ReaderClient) Seek(offset int64, whence int) (new int64, err error) {
	if err = r.conn.Encode(proto.ReaderSeek, offset, int8(whence)); err != nil {
		return
	}
	if err = r.conn.Decode(&new); err != nil {
		return
	}
	return
}

func (r *ReaderClient) Close() (err error) {
	if err = r.conn.Encode(proto.ReaderClose); err != nil {
		return
	}
	err = r.conn.Close()
	return
}

func (r *ReaderClient) Info() (info *storage.ReaderInfo) {
	i := storage.ReaderInfo{}
	if err := r.conn.Encode(proto.ReaderInfo); err != nil {
		return
	}
	if err := json.NewDecoder(r.conn).Decode(&i); err != nil {
		return
	}
	info = &i
	return
}
