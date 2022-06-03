package store

import (
	"errors"
	"github.com/cryptopunkscc/astrald/cslq"
	"github.com/cryptopunkscc/astrald/data"
	"github.com/cryptopunkscc/astrald/proto/block"
	"io"
)

var _ Store = &Binding{}

type Binding struct {
	t io.ReadWriter
}

func Bind(transport io.ReadWriter) *Binding {
	return &Binding{t: transport}
}

func (b *Binding) Open(blockID data.ID, flags uint32) (block.Block, error) {
	if err := cslq.Encode(b.t, "cvl", cmdOpen, blockID, flags); err != nil {
		return nil, err
	}

	var errCode int
	if err := cslq.Decode(b.t, "c", &errCode); err != nil {
		return nil, err
	}

	switch errCode {
	case success:
		return block.Bind(b.t), nil
	case errNotFound:
		return nil, errors.New("not found")
	case errUnavailable:
		return nil, errors.New("open unavailable")
	default:
		return nil, errors.New("protocol error: unknown error code")
	}
}

func (b *Binding) Create(alloc uint64) (block.Block, string, error) {
	if err := cslq.Encode(b.t, "cq", cmdCreate, alloc); err != nil {
		return nil, "", err
	}

	var errCode int
	var tempID string
	if err := cslq.Decode(b.t, "c [c]c", &errCode, &tempID); err != nil {
		return nil, "", err
	}

	switch errCode {
	case success:
		return block.Bind(b.t), tempID, nil
	case errFailed:
		return nil, "", errors.New("create failed")
	case errUnavailable:
		return nil, "", errors.New("create unavailable")
	default:
		return nil, "", errors.New("protocol error: unknown error code")
	}
}
