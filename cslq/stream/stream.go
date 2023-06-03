package stream

import (
	"errors"
	"github.com/cryptopunkscc/astrald/cslq"
	"io"
)

const errorCSLQ = "s"

// Stream wraps an io.ReadWriter and provides helper methods to write/read structs and errors.
// If io.ReadWriter is also an io.Closer, the Close() method will close it.
type Stream struct {
	ErrorSpace
	io.ReadWriter

	// CSLQ string used to encode error codes. Default is "s".
	ErrorType string
}

// NewStream retruns a stream reader/writer that uses the provided ErrorSpace for error encoding.
func NewStream(io io.ReadWriter, errorSpace ErrorSpace) *Stream {
	return &Stream{
		ErrorSpace: errorSpace,
		ReadWriter: io,
		ErrorType:  errorCSLQ,
	}
}

// Encode encodes and writes one or many structs to the stream.
func (stream *Stream) Encode(v ...interface{}) error {
	for _, i := range v {
		if err := cslq.Encode(stream, "v", i); err != nil {
			return err
		}
	}
	return nil
}

// Decode reads and decodes a struct from the stream. Argument v must be a pointer to a struct.
func (stream *Stream) Decode(v interface{}) error {
	return cslq.Decode(stream, "v", v)
}

// WriteError writes a ProtocolError (or nil) to the stream.
func (stream *Stream) WriteError(err *ProtocolError) error {
	if err == nil {
		return cslq.Encode(stream, stream.ErrorType, 0)
	}
	return cslq.Encode(stream, stream.ErrorType, err.ErrorCode())
}

// ReadError reads an error from the stream. On success, it returns nil or a ProtocolError. Other types of errors
// are actual errors.
func (stream *Stream) ReadError() error {
	var code int
	if err := cslq.Decode(stream, stream.ErrorType, &code); err != nil {
		return err
	}
	if code == 0 {
		return nil
	}
	if err, found := stream.ByCode(code); found {
		return err
	}
	return errors.New("invalid error code")
}

// Close closes the stream if the underlying transport is an io.Closer.
func (stream *Stream) Close() error {
	if closer, ok := stream.ReadWriter.(io.Closer); ok {
		return closer.Close()
	}
	return errors.New("close unsupported")
}
