package store

import (
	"errors"
	"github.com/cryptopunkscc/astrald/cslq"
	"github.com/cryptopunkscc/astrald/cslq/rpc"
	"github.com/cryptopunkscc/astrald/data"
	"github.com/cryptopunkscc/astrald/proto/block"
	"io"
)

func Serve(conn io.ReadWriter, store Store) (err error) {
	d := dispatcher{
		conn:  conn,
		Endec: cslq.NewEndec(conn),
		store: store,
	}
	err = d.dispatch()
	if errors.Is(err, errEnded) {
		err = nil
	}
	return
}

type dispatcher struct {
	conn io.ReadWriter
	*cslq.Endec
	store Store
}

func (d dispatcher) dispatch() error {
	return rpc.Dispatch(d.conn, "c", func(cmd uint8) error {
		switch cmd {
		case cmdOpen:
			return rpc.Dispatch(d.conn, "vl", d.open)

		case cmdCreate:
			return rpc.Dispatch(d.conn, "q", d.create)

		case cmdDownload:
			return rpc.Dispatch(d.conn, "vqq", d.download)

		case cmdEnd:
			if err := d.Encode("c", 0); err != nil {
				return err
			}
			return errEnded

		default:
			return errors.New("protocol violation: unknown command")
		}
	})
}

func (d *dispatcher) open(blockID data.ID, flags uint32) error {
	object, err := d.store.Open(blockID, flags)

	switch {
	case err == nil:
		if err := d.Encode("c", success); err != nil {
			return err
		}
		return block.Serve(d.conn, block.Wrap(object))

	default:
		return d.Encode("c", errNotFound)
	}

}

func (d *dispatcher) create(alloc uint64) error {
	_block, tempID, err := d.store.Create(alloc)

	switch {
	case err == nil:
		if err := d.Encode("c [c]c", success, tempID); err != nil {
			return err
		}
		return block.Serve(d.conn, block.Wrap(_block))

	default:
		return d.Encode("c [c]c", errFailed, "")
	}
}

func (d *dispatcher) download(blockID data.ID, offset uint64, limit uint64) error {
	rc, err := d.store.Download(blockID, offset, limit)

	switch {
	case err == nil:
		if err := d.Encode("c", success); err != nil {
			return err
		}
		defer rc.Close()
		_, err := io.Copy(d.conn, rc)
		return err

	default:
		return d.Encode("c", errNotFound)
	}
}
