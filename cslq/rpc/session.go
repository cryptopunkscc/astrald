package rpc

import (
	"errors"
	"github.com/cryptopunkscc/astrald/cslq"
	"io"
)

const DefaultErrorFormat = "s"

// Session wraps an io.ReadWriter and provides helper methods to write/read structs and errors.
// If io.ReadWriter is also an io.Closer, the Close() method will close it.
type Session[MethodType any] struct {
	ErrorSpace
	*cslq.Endec

	t io.ReadWriter
	// CSLQ string used to encode error codes. Default is DefaultErrorFormat.
	ErrorType string
}

// NewSession retruns a stream reader/writer that uses the provided ErrorSpace for error encoding.
func NewSession[MethodType any](t io.ReadWriter, errorSpace ErrorSpace) *Session[MethodType] {
	return &Session[MethodType]{
		ErrorSpace: errorSpace,
		t:          t,
		Endec:      cslq.NewEndec(t),
		ErrorType:  DefaultErrorFormat,
	}
}

// Invoke calls the method with provided argument. If res is not nil, it will also decode the result into res.
func (s *Session[T]) Invoke(method T, arg any, res any) error {
	if err := s.Encode(method); err != nil {
		return err
	}
	if arg != nil {
		if err := s.Encode(arg); err != nil {
			return err
		}
	}
	if err := s.DecodeErr(); err != nil {
		return err
	}
	if res != nil {
		return s.Decode(res)
	}
	return nil
}

// EncodeErr writes a RPCError (or nil) to the stream.
func (s *Session[_]) EncodeErr(err *RPCError) error {
	if err == nil {
		return s.Encodef(s.ErrorType, 0)
	}
	return s.Encodef(s.ErrorType, err.ErrorCode())
}

// DecodeErr reads an error from the stream. On success, it returns nil or a RPCError. Other types of errors
// are actual errors.
func (s *Session[_]) DecodeErr() error {
	var code int
	if err := s.Decodef(s.ErrorType, &code); err != nil {
		return err
	}
	if code == 0 {
		return nil
	}
	if err, found := s.ByCode(code); found {
		return err
	}
	return errors.New("invalid error code")
}

func (s *Session[_]) Transport() io.ReadWriter {
	return s.t
}
