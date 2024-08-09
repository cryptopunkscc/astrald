package objects

import (
	"context"
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/core"
	"github.com/cryptopunkscc/astrald/lib/query"
	"github.com/cryptopunkscc/astrald/mod/objects"
	"github.com/cryptopunkscc/astrald/object"
	"io"
	"time"
)

var _ objects.Reader = &NetworkReader{}

type NetworkReader struct {
	mod      *Module
	objectID object.ID
	consumer *astral.Identity
	provider *astral.Identity

	pos int64
	io.ReadCloser
}

func (r *NetworkReader) Read(p []byte) (n int, err error) {
	n, err = r.ReadCloser.Read(p)
	r.pos += int64(n)
	return n, err
}

func (r *NetworkReader) Seek(offset int64, whence int) (int64, error) {
	r.ReadCloser.Close()

	params := core.Params{
		"id": r.objectID.String(),
	}

	var o int64

	switch whence {
	case io.SeekStart:
		o = offset
	case io.SeekCurrent:
		o = r.pos + offset
	case io.SeekEnd:
		o = int64(r.objectID.Size) + offset
	}
	params.SetInt("offset", int(o))

	var q = astral.NewQuery(
		r.consumer,
		r.provider,
		core.Query(methodRead, params),
	)

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	conn, err := query.Route(ctx, r.mod.node, q)
	if err != nil {
		return 0, err
	}

	r.ReadCloser = conn
	r.pos = o

	return 0, nil
}

func (r *NetworkReader) Info() *objects.ReaderInfo {
	return &objects.ReaderInfo{Name: r.provider.String()}
}
