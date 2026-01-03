package channel

import (
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/sig"
)

// Channel is a bidirectional stream of astral objects.
type Channel struct {
	rw io.ReadWriter
	Receiver
	Sender
}

// New returns a new astral channel over the provided transport.
// To configure the channel, pass optional config functions:
//
//	New(rw, WithFormats("json", "text+"))
//
// See config.go for available config functions.
func New(rw io.ReadWriter, fn ...ConfigFunc) *Channel {
	var ch = &Channel{rw: rw}
	var cfg Config

	// apply config
	for _, f := range fn {
		f(&cfg)
	}

	ch.Receiver = newReceiver(rw, &cfg)
	ch.Sender = newSender(rw, &cfg)

	return ch
}

// NewReceiver returns a new receive-only channel.
func NewReceiver(r io.Reader, fn ...ConfigFunc) Receiver {
	var cfg Config
	for _, f := range fn {
		f(&cfg)
	}
	return newReceiver(r, &cfg)
}

// NewSender returns a new send-only channel.
func NewSender(w io.Writer, fn ...ConfigFunc) Sender {
	var cfg Config
	for _, f := range fn {
		f(&cfg)
	}
	return newSender(w, &cfg)
}

func newReceiver(r io.Reader, cfg *Config) Receiver {
	// build the channel
	switch cfg.fmtIn {
	case "", Binary:
		return NewBinaryReceiver(r)
	case JSON:
		return NewJSONReceiver(r)
	case Text, TextTyped:
		return NewTextReceiver(r)
	default:
		return NewReceiverError(fmt.Errorf("unsupported input format: %s", cfg.fmtIn))
	}
}

func newSender(w io.Writer, cfg *Config) Sender {
	switch cfg.fmtOut {
	case "", Binary:
		return NewBinarySender(w)
	case JSON:
		return NewJSONSender(w)
	case Text, TextTyped:
		return NewTextSender(w, strings.HasSuffix(cfg.fmtOut, "+"))
	case Render:
		return NewRenderSender(w)
	default:
		return NewSenderError(fmt.Errorf("unsupported output format: %s", cfg.fmtOut))
	}
}

// Collect receives objects from the channel until EOF and passes them to the collector. It stops when
// the collector returns an error or when Receive() returns a non-EOF error.
func (ch Channel) Collect(collector func(astral.Object) error) error {
	for {
		o, err := ch.Receive()
		switch {
		case err == nil:
			if err = collector(o); err != nil {
				return err
			}

		case errors.Is(err, io.EOF):
			return nil

		default:
			return err
		}
	}
}

// Handle receives objects from the channel until EOF and passes them to the handler. It closes the channel
// and returns when the context is canceled.
func (ch Channel) Handle(ctx *astral.Context, handler func(astral.Object)) error {
	done := make(chan struct{})
	defer close(done)

	var retErr sig.Value[error]

	go func() {
		select {
		case <-done:
		case <-ctx.Done():
			retErr.Swap(nil, ctx.Err())
			ch.Close()
		}
	}()

	for {
		o, err := ch.Receive()
		switch {
		case err == nil:
			handler(o)

		case errors.Is(err, io.EOF):
			return nil

		default:
			err, _ = retErr.Swap(nil, err)
			return err
		}
	}
}

// Close closes the channel if the transport supports it.
func (ch Channel) Close() error {
	if c, ok := ch.rw.(io.Closer); ok {
		return c.Close()
	}
	return ErrCloseUnsupported
}

// Transport returns the underlying transport.
func (ch Channel) Transport() io.ReadWriter {
	return ch.rw
}
