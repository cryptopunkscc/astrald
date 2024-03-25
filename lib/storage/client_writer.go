package storage

import (
	"github.com/cryptopunkscc/astrald/data"
	"github.com/cryptopunkscc/astrald/mod/storage"
	proto "github.com/cryptopunkscc/astrald/mod/storage/srv"
	"io"
)

type WriterClient struct {
	conn proto.BinaryEncoderCloser
}

func NewWriterClient(conn io.ReadWriteCloser) storage.Writer {
	return &WriterClient{proto.NewBinaryEncoderCloser(conn)}
}

func (w WriterClient) Write(p []byte) (n int, err error) {
	if err = w.conn.Encode(proto.WriterWrite, int32(len(p)), p); err != nil {
		return
	}
	var n32 int32
	if err = w.conn.Decode(&n32); err != nil {
		return
	}
	n = int(n32)
	return
}

func (w WriterClient) Discard() (err error) {
	if err = w.conn.Encode(proto.WriterDiscard); err != nil {
		return
	}
	err = w.conn.Close()
	return
}

func (w WriterClient) Commit() (id data.ID, err error) {
	if err = w.conn.Encode(proto.WriterCommit); err != nil {
		return
	}
	if err = w.conn.Decode(&id); err != nil {
		return
	}
	err = w.conn.Close()
	return
}
