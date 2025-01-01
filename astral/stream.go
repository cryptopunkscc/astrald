package astral

import (
	"bytes"
	"errors"
	"io"
)

var _ HasBlueprints = &Stream{}
var _ ObjectWriter = &Stream{}
var _ ObjectReader = &Stream{}

type Stream struct {
	blueprints Blueprints
	rw         io.ReadWriter
}

func NewStream(rw io.ReadWriter, bp *Blueprints) *Stream {
	s := &Stream{
		rw: rw,
	}
	s.blueprints.Parent = bp
	return s
}

// ReadObject reads the next object from the stream
func (stream *Stream) ReadObject() (object Object, n int64, err error) {
	var buf []byte

	// read the buffer
	n, err = (*Bytes32)(&buf).ReadFrom(stream.rw)
	if err != nil {
		return
	}

	// read an object from the buffer
	object, _, err = stream.blueprints.Read(bytes.NewReader(buf), false)

	return
}

// WriteObject writes an object to the stream
func (stream *Stream) WriteObject(object Object) (n int64, err error) {
	var buf = &bytes.Buffer{}

	// buffer the type
	_, err = String8(object.ObjectType()).WriteTo(buf)
	if err != nil {
		return
	}

	// buffer the payload
	_, err = object.WriteTo(buf)
	if err != nil {
		return
	}

	if buf.Len() == 1 {
		// skip empty payload untyped objects
		return 0, nil
	}

	return Bytes32(buf.Bytes()).WriteTo(stream.rw) // write the buffer
}

// Blueprints returns Stream's blueprints. Streams have their own blueprints that inherit from the provided parent.
func (stream *Stream) Blueprints() *Blueprints {
	return &stream.blueprints
}

// Close tries to invoke rw's Close
func (stream *Stream) Close() error {
	if c, ok := stream.rw.(io.Closer); ok {
		return c.Close()
	}
	return errors.New("transport does not implement io.Closer")
}
