package store

import (
	"errors"
	"github.com/cryptopunkscc/astrald/cslq"
	"github.com/cryptopunkscc/astrald/cslq/rpc"
	"github.com/cryptopunkscc/astrald/data"
	"github.com/cryptopunkscc/astrald/proto/block"
	"io"
)

func Serve(conn io.ReadWriter, store Store) error {
	var d = &dispatcher{conn: conn, store: store}
	for {
		if err := d.dispatch(); err != nil {
			if errors.Is(err, errEnded) {
				return nil
			}
			return err
		}
	}
}

type dispatcher struct {
	conn  io.ReadWriter
	store Store
}

func (d *dispatcher) dispatch() error {
	return rpc.Dispatch(d.conn, "c", func(cmd uint8) error {
		switch cmd {
		case cmdOpen:
			return rpc.Dispatch(d.conn, "vl", d.open)

		case cmdCreate:
			return rpc.Dispatch(d.conn, "q", d.create)

		case cmdEnd:
			if err := cslq.Encode(d.conn, "c", 0); err != nil {
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
		if err := cslq.Encode(d.conn, "c", success); err != nil {
			return err
		}
		return block.Serve(d.conn, block.Wrap(object))

	default:
		return cslq.Encode(d.conn, "c", errNotFound)
	}

}

func (d *dispatcher) create(alloc uint64) error {
	_block, tempID, err := d.store.Create(alloc)

	switch {
	case err == nil:
		if err := cslq.Encode(d.conn, "c [c]c", success, tempID); err != nil {
			return err
		}
		return block.Serve(d.conn, block.Wrap(_block))

	default:
		return cslq.Encode(d.conn, "c [c]c", errFailed, "")
	}
}
