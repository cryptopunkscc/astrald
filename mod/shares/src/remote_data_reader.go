package shares

import (
	"context"
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/data"
	"github.com/cryptopunkscc/astrald/mod/storage"
	"github.com/cryptopunkscc/astrald/net"
	"github.com/cryptopunkscc/astrald/node/router"
	"io"
	"time"
)

var _ storage.Reader = &RemoteDataReader{}

type RemoteDataReader struct {
	mod    *Module
	caller id.Identity
	target id.Identity
	dataID data.ID
	pos    int64
	io.ReadCloser
}

func (r *RemoteDataReader) Read(p []byte) (n int, err error) {
	n, err = r.ReadCloser.Read(p)
	r.pos += int64(n)
	return n, err
}

func (r *RemoteDataReader) Seek(offset int64, whence int) (int64, error) {
	r.ReadCloser.Close()

	params := router.Params{
		"id": r.dataID.String(),
	}

	var o int64

	switch whence {
	case io.SeekStart:
		o = offset
	case io.SeekCurrent:
		o = r.pos + offset
	case io.SeekEnd:
		o = int64(r.dataID.Size) + offset
	}
	params.SetInt("offset", int(o))

	var query = router.Query(readServiceName, params)

	var q = net.NewQuery(
		r.caller,
		r.target,
		query,
	)

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	conn, err := net.Route(ctx, r.mod.node.Router(), q)
	if err != nil {
		return 0, err
	}

	r.ReadCloser = conn
	r.pos = o

	return 0, nil
}

func (r *RemoteDataReader) Info() *storage.ReaderInfo {
	return &storage.ReaderInfo{Name: "shares"}
}
