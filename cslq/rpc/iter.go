package rpc

import (
	"errors"
	"github.com/cryptopunkscc/astrald/cslq"
	"io"
	"reflect"
)

// Iter expects fn to have a single argument of a type that can be decoded using cslq. Iter will decode values of that
// type from the reader and call the provided function. If the function returns exactly one bool, if the bool is true,
// iter will stop. Otherwise, returned values are ignored.
func Iter(r io.Reader, fn interface{}) error {
	var rv = reflect.ValueOf(fn)

	if rv.Kind() != reflect.Func {
		return errors.New("not a func")
	}

	if rv.Type().NumIn() != 1 {
		return errors.New("fn has to be a one arg func")
	}

	var itemType = rv.Type().In(0)

	format, err := cslq.Compile(typeToPattern(itemType))
	if err != nil {
		return err
	}

	for {
		var item = reflect.New(itemType)

		if err := format.Decode(r, item.Interface()); err != nil {
			return err
		}

		var ret = rv.Call([]reflect.Value{item.Elem()})

		if len(ret) == 1 && ret[0].Kind() == reflect.Bool {
			if ret[0].Bool() {
				return nil
			}
		}
	}
}
