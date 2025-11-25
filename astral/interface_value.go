package astral

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"reflect"
)

type interfaceValue struct {
	reflect.Value
}

var _ Object = &interfaceValue{}

func (i interfaceValue) ObjectType() string {
	return ""
}

func (i interfaceValue) WriteTo(w io.Writer) (n int64, err error) {
	if i.IsNil() {
		return Bytes32{}.WriteTo(w)
	}

	var buf = &bytes.Buffer{}

	typer, ok := i.Interface().(ObjectTyper)
	if !ok {
		return n, errors.New("interface type not supported")
	}

	_, err = String8(typer.ObjectType()).WriteTo(buf)
	if err != nil {
		return
	}

	wto, ok := i.Interface().(io.WriterTo)
	if !ok {
		return n, errors.New("interface type not supported")
	}

	_, err = wto.WriteTo(buf)

	return Bytes32(buf.Bytes()).WriteTo(w)
}

func (i interfaceValue) ReadFrom(r io.Reader) (n int64, err error) {
	var buf Bytes32
	var o Object

	n, err = buf.ReadFrom(r)
	if err != nil {
		return
	}

	if len(buf) == 0 {
		i.Set(reflect.Zero(i.Type()))
		return
	}

	o, _, err = ExtractBlueprints(r).Read(bytes.NewReader(buf))
	if err != nil {
		return
	}

	var ov = reflect.ValueOf(o)
	if !ov.CanConvert(i.Type()) {
		err = fmt.Errorf("cannot convert %s to %s", ov.Type(), i.Type())
		return
	}

	i.Set(ov.Convert(i.Type()))

	return
}
