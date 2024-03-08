package storage

import (
	"github.com/cryptopunkscc/astrald/cslq"
	"github.com/cryptopunkscc/astrald/data"
	"github.com/cryptopunkscc/astrald/mod/storage"
	"io"
)

const (
	WriterWrite = byte(iota) + 1
	WriterCommit
	WriterDiscard
)

type writerClient struct {
	conn io.ReadWriteCloser
}

func newWriterClient(conn io.ReadWriteCloser) storage.Writer {
	return &writerClient{conn}
}

func (w writerClient) Write(p []byte) (n int, err error) {
	err = cslq.Encode(w.conn, "c [l]c", WriterWrite, p)
	if err != nil {
		n = -1
		return
	}
	cslq.Decode(w.conn, "l", &n)
	return
}

func (w writerClient) Commit() (id data.ID, err error) {
	if err = cslq.Encode(w.conn, "c", WriterCommit); err != nil {
		return
	}
	if err = cslq.Decode(w.conn, "v", &id); err != nil {
		return
	}
	err = w.conn.Close()
	return
}

func (w writerClient) Discard() (err error) {
	if err = cslq.Encode(w.conn, "c", WriterDiscard); err != nil {
		return
	}
	err = w.conn.Close()
	return
}
