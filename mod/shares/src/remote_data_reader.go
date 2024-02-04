package shares

import (
	"context"
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/data"
	"github.com/cryptopunkscc/astrald/mod/storage"
	"github.com/cryptopunkscc/astrald/net"
	"github.com/cryptopunkscc/astrald/node/router"
	"io"
	"strconv"
	"time"
)

var _ storage.DataReader = &RemoteDataReader{}

type RemoteDataReader struct {
	mod    *Module
	dataID data.ID
	caller id.Identity
	target id.Identity
	io.ReadCloser
	pos int
}

func (r *RemoteDataReader) Read(p []byte) (n int, err error) {
	n, err = r.ReadCloser.Read(p)
	r.pos += n
	return n, err
}

func (r *RemoteDataReader) Seek(offset int64, whence int) (int64, error) {
	r.ReadCloser.Close()

	params := map[string]string{
		"id": r.dataID.String(),
	}

	var o uint64

	switch whence {
	case io.SeekStart:
		o = uint64(offset)
	case io.SeekCurrent:
		o = uint64(int64(r.pos) + offset)
	case io.SeekEnd:
		o = uint64(int64(r.dataID.Size) + offset)
	}
	params["offset"] = strconv.FormatUint(o, 10)

	var query = router.FormatQuery(readServiceName, params)

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
	r.pos = int(o)

	return 0, nil
}

func (r *RemoteDataReader) Info() *storage.ReaderInfo {
	return &storage.ReaderInfo{Name: "shares"}
}
