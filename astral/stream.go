package astral

import (
	"bytes"
	"errors"
	"github.com/cryptopunkscc/astrald/streams"
	"io"
)

var _ ObjectReader = &Stream{}
var _ HasBlueprints = &Stream{}

type Stream struct {
	StreamReader
	StreamWriter
}

var _ HasBlueprints = &StreamReader{}

type StreamReader struct {
	io.Reader
	Bits       int
	blueprints *Blueprints
}

type StreamWriter struct {
	io.Writer
	Bits int
}

func NewStream(rw io.ReadWriter, bits int, bp *Blueprints) *Stream {
	return &Stream{
		StreamReader: StreamReader{
			Reader:     rw,
			Bits:       bits,
			blueprints: bp,
		},
		StreamWriter: StreamWriter{
			Writer: rw,
			Bits:   bits,
		},
	}
}

func NewStreamReader(reader io.Reader, bits int, blueprints *Blueprints) *StreamReader {
	return &StreamReader{
		Reader:     reader,
		Bits:       bits,
		blueprints: blueprints,
	}
}

func NewStreamWriter(writer io.Writer, bits int) *StreamWriter {
	return &StreamWriter{Writer: writer, Bits: bits}
}

func (s *StreamWriter) WriteObject(objects ...Object) (n int64, err error) {
	var m int64
	for _, object := range objects {
		m, err = s.writeObject(object)
		n += m
		if err != nil {
			return
		}
	}
	return
}

func (s *StreamWriter) writeObject(o Object) (n int64, err error) {
	var buf = &bytes.Buffer{}

	_, err = streams.WriteAllTo(buf, String8(o.ObjectType()), o)
	if err != nil {
		return
	}

	switch s.Bits {
	case 0:
		return Bytes(buf.Bytes()).WriteTo(s)
	case 8:
		return Bytes8(buf.Bytes()).WriteTo(s)
	case 16:
		return Bytes16(buf.Bytes()).WriteTo(s)
	case 32:
		return Bytes32(buf.Bytes()).WriteTo(s)
	case 64:
		return Bytes64(buf.Bytes()).WriteTo(s)
	default:
		return 0, errors.New("unsupported width")
	}
}

func (s *StreamReader) ReadObject() (o Object, n int64, err error) {
	if s.Bits == 0 {
		var m int64
		// read the type of the object
		var typeName String8
		n, err = typeName.ReadFrom(s)
		if err != nil {
			return
		}

		if s.blueprints != nil {
			o = s.blueprints.Make(string(typeName))
		}
		if o == nil {
			o = &RawObject{Type: string(typeName)}
		}

		m, err = o.ReadFrom(s)
		n += m
		return
	}

	var r io.Reader

	switch s.Bits {
	case 8:
		var frame Bytes8
		n, err = frame.ReadFrom(s)
		r = bytes.NewReader(frame)
	case 16:
		var frame Bytes16
		n, err = frame.ReadFrom(s)
		r = bytes.NewReader(frame)
	case 32:
		var frame Bytes32
		n, err = frame.ReadFrom(s)
		r = bytes.NewReader(frame)
	case 64:
		var frame Bytes64
		n, err = frame.ReadFrom(s)
		r = bytes.NewReader(frame)
	default:
		err = errors.New("unsupported width")
	}
	if err != nil {
		return
	}

	o, n, err = s.Blueprints().Read(r, false)

	//TODO: should leftover data be considered an error?
	return
}

func (s *StreamReader) Blueprints() *Blueprints {
	return s.blueprints
}
