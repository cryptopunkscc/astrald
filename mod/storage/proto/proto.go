package proto

import (
	"github.com/cryptopunkscc/astrald/cslq/stream"
	"io"
)

var es stream.ErrorSpace

var (
	// ErrUnavailable - requested data is not available
	ErrUnavailable = es.NewError(0x01, "unavailable")

	// ErrSeekUnsupported - seek operation is not supported
	ErrSeekUnsupported = es.NewError(0x02, "seek unsupported")

	// ErrSizeMismatch - size from data.ID and actual data size differ
	ErrSizeMismatch = es.NewError(0x03, "block size mismatch")

	// ErrInvalidOffset - offset is less than zero or more than the data length
	ErrInvalidOffset = es.NewError(0x04, "invalid offset")

	// ErrInvalidLength - length exceeds the remaining data or lenght is less than 1
	ErrInvalidLength = es.NewError(0x05, "invalid length")
)

func New(rw io.ReadWriter) *stream.Stream {
	var s = stream.NewStream(rw, es)
	s.ErrorType = "c"
	return s
}
