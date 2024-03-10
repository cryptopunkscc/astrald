package storage

import (
	"github.com/cryptopunkscc/astrald/cslq"
	"github.com/cryptopunkscc/astrald/mod/storage"
	api "github.com/cryptopunkscc/astrald/mod/storage/cslq"
	"io"
)

type readerClient struct {
	conn io.ReadWriteCloser
}

func newReaderClient(conn io.ReadWriteCloser) storage.Reader {
	return &readerClient{conn: conn}
}

func (r *readerClient) Read(p []byte) (n int, err error) {
	if err = cslq.Encode(r.conn, "c l", api.ReaderRead, cap(p)); err != nil {
		return
	}
	n, err = r.conn.Read(p)
	return
}

func (r *readerClient) Seek(offset int64, whence int) (new int64, err error) {
	if err = cslq.Encode(r.conn, "c q l", api.ReaderSeek, offset, whence); err != nil {
		return
	}
	if err = cslq.Decode(r.conn, "q", &new); err != nil {
		return
	}
	return
}

func (r *readerClient) Close() (err error) {
	if err = cslq.Encode(r.conn, "c", api.ReaderClose); err != nil {
		return
	}
	err = r.conn.Close()
	return
}

func (r *readerClient) Info() (info *storage.ReaderInfo) {
	i := storage.ReaderInfo{}
	if err := cslq.Encode(r.conn, "c", api.ReaderInfo); err != nil {
		return
	}
	if err := cslq.Decode(r.conn, "{[c]c}", &i); err != nil {
		return
	}
	info = &i
	return
}
