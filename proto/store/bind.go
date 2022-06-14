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
	*cslq.Endec
	conn io.ReadWriteCloser
}

func Bind(conn io.ReadWriteCloser) *Binding {
	return &Binding{
		Endec: cslq.NewEndec(conn),
		conn:  conn,
	}
}

func (b *Binding) Open(blockID data.ID, flags uint32) (block.Block, error) {
	if err := b.Encode("cvl", cmdOpen, blockID, flags); err != nil {
		return nil, err
	}

	var errCode int
	if err := b.Decode("c", &errCode); err != nil {
		return nil, err
	}

	switch errCode {
	case success:
		return block.Bind(b.conn), nil

	case errNotFound:
		return nil, errors.New("not found")

	case errUnavailable:
		return nil, errors.New("open unavailable")

	default:
		return nil, errors.New("protocol error: unknown error code")
	}
}

func (b *Binding) Create(alloc uint64) (block.Block, string, error) {
	if err := b.Encode("cq", cmdCreate, alloc); err != nil {
		return nil, "", err
	}

	var errCode int
	var tempID string
	if err := b.Decode("c [c]c", &errCode, &tempID); err != nil {
		return nil, "", err
	}

	switch errCode {
	case success:
		return block.Bind(b.conn), tempID, nil

	case errFailed:
		return nil, "", errors.New("create failed")

	case errUnavailable:
		return nil, "", errors.New("create unavailable")

	default:
		return nil, "", errors.New("protocol error: unknown error code")
	}
}

func (b *Binding) Download(blockID data.ID, offset uint64, limit uint64) (io.ReadCloser, error) {
	if err := b.Encode("cvqq", cmdDownload, blockID, offset, limit); err != nil {
		return nil, err
	}

	var errCode int
	if err := b.Decode("c", &errCode); err != nil {
		return nil, err
	}

	if errCode == 0 {
		return b.conn, nil
	}

	defer b.conn.Close()
	switch errCode {
	case errNotFound:
		return nil, ErrNotFound

	default:
		return nil, errors.New("protocol error: unknown error code")
	}
}
