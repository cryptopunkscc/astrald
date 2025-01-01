package astral

import (
	"bytes"
	"io"
	"reflect"
)

var _ Object = &Box[Object]{}

// Box is an anonymous object that adds 32-bit prefix length encoding to its content.
type Box[T Object] struct {
	Content *T
}

func NewBox[T Object](content *T) *Box[T] {
	return &Box[T]{Content: content}
}

func (b Box[T]) ObjectType() string {
	return ""
}

func (b Box[T]) WriteTo(w io.Writer) (n int64, err error) {
	var buf = &bytes.Buffer{}

	// write the payload to the buffer
	_, err = (*b.Content).WriteTo(buf)
	if err != nil {
		return
	}

	return Bytes32(buf.Bytes()).WriteTo(w)
}

func (b *Box[T]) ReadFrom(r io.Reader) (n int64, err error) {
	var buf []byte

	n, err = (*Bytes32)(&buf).ReadFrom(r)
	if err != nil {
		return
	}

	typ := reflect.TypeOf(b.Content).Elem()
	if typ.Kind() == reflect.Pointer {
		*b.Content = reflect.New(typ.Elem()).Interface().(T)
	}

	_, err = (*b.Content).ReadFrom(bytes.NewReader(buf))

	return
}
