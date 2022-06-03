package rpc

import (
	"errors"
	"github.com/cryptopunkscc/astrald/cslq"
	"io"
	"reflect"
)

// Dispatch decodes arguments for the function using the provided pattern and calls the function. The function has
// to return exactly one error value.
func Dispatch(r io.Reader, pattern string, fn interface{}) (err error) {
	var rv = reflect.ValueOf(fn)

	if rv.Kind() != reflect.Func {
		return errors.New("argument is not a func")
	}

	format, err := cslq.Compile(pattern)
	if err != nil {
		return err
	}

	var args = make([]reflect.Value, 0)
	var ptrs = make([]interface{}, 0)

	for i := 0; i < rv.Type().NumIn(); i++ {
		var argType = rv.Type().In(i)
		var arg = reflect.New(argType)
		args = append(args, arg.Elem())
		ptrs = append(ptrs, arg.Interface())
	}

	if err := format.Decode(r, ptrs...); err != nil {
		return err
	}

	var ret = rv.Call(args)

	errVal := reflect.ValueOf(&err)
	if ret[0].IsZero() {
		return nil
	}

	if !ret[0].CanConvert(errVal.Elem().Type()) {
		return errors.New("return value not an error type")
	}
	errVal.Elem().Set(ret[0])

	return err
}
