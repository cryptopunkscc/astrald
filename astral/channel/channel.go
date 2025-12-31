package channel

import (
	"fmt"
	"io"
	"strings"
)

// Channel is a bidirectional stream of astral objects.
type Channel struct {
	rw io.ReadWriter
	Reader
	Writer
}

// New returns a new astral channel over the provided transport.
// To configure the channel, pass optional config functions:
//
//	New(rw, InFmt("json"), OutFmt("text+"))
//
// See config.go for available config functions.
func New(rw io.ReadWriter, fn ...configFunc) *Channel {
	var ch = &Channel{rw: rw}
	var cfg channelConfig

	// apply config
	for _, f := range fn {
		f(&cfg)
	}

	ch.Reader = newReader(rw, &cfg)
	ch.Writer = newWriter(rw, &cfg)

	return ch
}

// NewReader returns a new read-only channel.
func NewReader(r io.Reader, fn ...configFunc) Reader {
	var cfg channelConfig
	for _, f := range fn {
		f(&cfg)
	}
	return newReader(r, &cfg)
}

// NewWriter returns a new write-only channel.
func NewWriter(w io.Writer, fn ...configFunc) Writer {
	var cfg channelConfig
	for _, f := range fn {
		f(&cfg)
	}
	return newWriter(w, &cfg)
}

func newReader(r io.Reader, cfg *channelConfig) Reader {
	// build the channel
	switch cfg.fmtIn {
	case "", Binary:
		return NewBinaryReader(r)
	case JSON:
		return NewJSONReader(r)
	case Text, TextTyped:
		return NewTextReader(r)
	default:
		return NewReaderError(fmt.Errorf("unsupported input format: %s", cfg.fmtIn))
	}
}

func newWriter(w io.Writer, cfg *channelConfig) Writer {
	switch cfg.fmtOut {
	case "", Binary:
		return NewBinaryWriter(w)
	case JSON:
		return NewJSONWriter(w)
	case Text, TextTyped:
		return NewTextWriter(w, strings.HasSuffix(cfg.fmtOut, "+"))
	default:
		return NewWriterError(fmt.Errorf("unsupported output format: %s", cfg.fmtOut))
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
