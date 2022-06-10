package block

import (
	"errors"
	"github.com/cryptopunkscc/astrald/cslq"
	"github.com/cryptopunkscc/astrald/cslq/rpc"
	"io"
)

type dispatcher struct {
	t     io.ReadWriter
	block Block
}

// Serve serves the provided Block over the transport using io:block protocol
func Serve(t io.ReadWriter, block Block) error {
	var d = &dispatcher{t: t, block: block}

	for {
		if err := rpc.Dispatch(d.t, "c", d.dispatch); err != nil {
			if errors.Is(err, ErrEnded) {
				return nil
			}
			block.Close()
			return err
		}
	}
}

func (d *dispatcher) dispatch(cmd byte) error {
	switch cmd {
	case cmdRead:
		return rpc.Dispatch(d.t, "s", d.read)

	case cmdWrite:
		return rpc.Dispatch(d.t, "[s]c", d.write)

	case cmdSeek:
		return rpc.Dispatch(d.t, "q c", d.seek)

	case cmdFinalize:
		return d.finalize()

	case cmdClose:
		if err := d.close(); err != nil {
			return err
		}
		return ErrEnded

	default:
		return ProtocolError{"unknown command"}
	}
}

func (d *dispatcher) read(maxBytes int) error {
	var errCode int
	var buf = make([]byte, maxBytes)

	n, err := d.block.Read(buf)

	switch {
	case err == nil:
	case errors.Is(err, ErrUnavailable):
		errCode = errUnavailable
	case errors.Is(err, io.EOF):
		errCode = errEOB
	default:
		errCode = errFailed
	}

	return cslq.Encode(d.t, "c [s]c", errCode, buf[:n])
}

func (d *dispatcher) write(buf []byte) error {
	var errCode int

	n, err := d.block.Write(buf)

	switch {
	case err == nil:
	case errors.Is(err, ErrUnavailable):
		errCode = errUnavailable

	default:
		errCode = errFailed
	}

	return cslq.Encode(d.t, "c s", errCode, n)
}

func (d *dispatcher) seek(pos int64, whence int) error {
	var errCode int

	n, err := d.block.Seek(pos, whence)

	switch {
	case err == nil:
	case errors.Is(err, ErrUnavailable):
		errCode = errUnavailable
	default:
		errCode = errFailed
	}

	return cslq.Encode(d.t, "c q", errCode, n)
}

func (d *dispatcher) finalize() error {
	var errCode int

	id, err := d.block.Finalize()

	switch {
	case err == nil:
	case errors.Is(err, ErrUnavailable):
		errCode = errUnavailable
	default:
		errCode = errFailed
	}

	return cslq.Encode(d.t, "c v", errCode, id)
}

func (d *dispatcher) close() error {
	if err := d.block.Close(); err != nil {
		return err
	}

	return cslq.Encode(d.t, "c", 0)
}
