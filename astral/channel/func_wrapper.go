package channel

import (
	"errors"
	"reflect"

	"github.com/cryptopunkscc/astrald/astral"
)

type funcWrapper struct {
	ObjectType string
	fn         reflect.Value
	argType    reflect.Type
	hasErr     bool
}

func wrapFunc(fnValue any) (*funcWrapper, error) {
	fn := reflect.ValueOf(fnValue)

	switch {
	case fn.Kind() != reflect.Func:
		return nil, errors.New("not a function")

	case fn.Type().NumIn() != 1:
		return nil, errors.New("function needs to take exactly one argument")
	case fn.Type().NumOut() > 1:
		return nil, errors.New("function can only return an error value")
	}

	argType := fn.Type().In(0)

	if !argType.Implements(objectType) && !argType.Implements(errorType) {
		return nil, errors.New("argument needs to be an astral.Object or error")
	}

	f := &funcWrapper{fn: fn, argType: argType}

	if fn.Type().NumOut() == 1 {
		if !fn.Type().Out(0).Implements(errorType) {
			return nil, errors.New("function can only return an error value")
		}
		f.hasErr = true
	}

	if argType.Kind() != reflect.Ptr {
		return f, nil
	}

	// get object type
	argVal := reflect.New(argType.Elem())
	obj := argVal.Interface().(astral.Object)
	f.ObjectType = obj.ObjectType()

	return f, nil
}

func (fn *funcWrapper) canCall(obj astral.Object) bool {
	return reflect.ValueOf(obj).Type().AssignableTo(fn.argType)
}

func (fn *funcWrapper) call(obj astral.Object) error {
	// create the value for the argument
	argVal := reflect.New(fn.argType).Elem()

	// set it to the object
	argVal.Set(reflect.ValueOf(obj))

	// call the wrapped function
	ret := fn.fn.Call([]reflect.Value{argVal})

	// if it returns an error, return it
	if fn.hasErr {
		if ret[0].IsNil() {
			return nil
		}
		return ret[0].Interface().(error)
	}

	return nil
}
