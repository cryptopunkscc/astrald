package block

import (
	"errors"
	"github.com/cryptopunkscc/astrald/cslq"
	"github.com/cryptopunkscc/astrald/data"
	"io"
	"math"
)

var _ Block = &Binding{}

type Binding struct {
	t io.ReadWriter
}

// Bind binds to a Block interface over the provided transport using the io:block protocol
func Bind(t io.ReadWriter) *Binding {
	return &Binding{t: t}
}

func (b *Binding) Read(p []byte) (n int, err error) {
	if len(p) > math.MaxUint16 {
		p = p[:math.MaxUint16]
	}

	// send request
	if err := cslq.Encode(b.t, "c s", cmdRead, len(p)); err != nil {
		return 0, err
	}

	// read the response
	var errCode int
	var buf []byte
	if err := cslq.Decode(b.t, "c [s]c", &errCode, &buf); err != nil {
		return 0, err
	}

	n = len(buf)
	copy(p[:n], buf[:n])

	// convert the error code
	switch errCode {
	case success:
	case errEOB:
		err = io.EOF
	case errFailed:
		err = ErrFailed
	case errUnavailable:
		err = ErrUnavailable
	default:
		err = ProtocolError{"invalid error code"}
	}

	return
}

func (b *Binding) Write(p []byte) (n int, err error) {
	// send request
	if err := cslq.Encode(b.t, "c [s]c", cmdWrite, p); err != nil {
		return 0, err
	}

	// read the response
	var errCode int
	if err := cslq.Decode(b.t, "c s", &errCode, &n); err != nil {
		return 0, err
	}

	// convert the error code
	switch errCode {
	case 0:
	case errNoSpace:
		err = errors.New("no space left on device")
	case errFailed:
		err = ErrFailed
	case errUnavailable:
		err = ErrUnavailable
	default:
		err = ProtocolError{"invalid error code"}
	}

	return
}

func (b *Binding) Seek(offset int64, whence int) (n int64, err error) {
	// send request
	if err := cslq.Encode(b.t, "c q c", cmdSeek, offset, whence); err != nil {
		return 0, err
	}

	// read the response
	var errCode int
	if err := cslq.Decode(b.t, "c q", &errCode, &n); err != nil {
		return 0, err
	}

	// convert the error code
	switch errCode {
	case 0:
	case errFailed:
		err = ErrFailed
	case errUnavailable:
		err = ErrUnavailable
	default:
		err = ProtocolError{"invalid error code"}
	}

	return
}

func (b *Binding) Finalize() (id data.ID, err error) {
	// send request
	if err := cslq.Encode(b.t, "c", cmdFinalize); err != nil {
		return data.ID{}, err
	}

	// read the response
	var errCode int
	if err := cslq.Decode(b.t, "c v", &errCode, &id); err != nil {
		return data.ID{}, err
	}

	// convert the error code
	switch errCode {
	case success:
	case errFailed:
		err = ErrFailed
	case errUnavailable:
		err = ErrUnavailable
	default:
		err = ProtocolError{"invalid error code"}
	}

	return
}

func (b *Binding) End() error {
	if err := cslq.Encode(b.t, "c", cmdEnd); err != nil {
		return err
	}

	var null byte

	if err := cslq.Decode(b.t, "c", &null); err != nil {
		return err
	}

	if null != 0 {
		return ProtocolError{"final byte not null"}
	}

	return nil
}
