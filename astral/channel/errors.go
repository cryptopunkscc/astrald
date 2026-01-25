package channel

import (
	"errors"

	"github.com/cryptopunkscc/astrald/astral"
)

var (
	ErrCloseUnsupported = errors.New("transport doesn't support closing")
	ErrTextUnsupported  = errors.New("the object does not support text marshaling")
	ErrBreak            = errors.New("break the switch")
)

// ReceiverError implements the Receiver interface. Its Receive() method always returns the wrapped error.
type ReceiverError struct {
	err error
}

func NewReceiverError(err error) *ReceiverError {
	return &ReceiverError{err: err}
}

var _ Receiver = &ReceiverError{}

func (r ReceiverError) Receive() (astral.Object, error) {
	return nil, r.err
}

// SenderError implements the Sender interface. Its Send() method always returns the wrapped error.
type SenderError struct {
	err error
}

var _ Sender = &SenderError{}

func NewSenderError(err error) *SenderError {
	return &SenderError{err: err}
}

func (w SenderError) Send(astral.Object) error {
	return w.err
}
