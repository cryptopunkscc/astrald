package rpc

import (
	"errors"
	"fmt"
	"github.com/cryptopunkscc/astrald/cslq"
	"io"
	"reflect"
)

func ImportFuncAs(rw io.ReadWriter, funcPtr interface{}, funcName string) error {
	var rv = reflect.ValueOf(funcPtr)
	if (rv.Kind() != reflect.Ptr) || (rv.Elem().Kind() != reflect.Func) {
		fmt.Println(rv.Type())
		return errors.New("v is not a func pointer")
	}

	var (
		fn        = wrapFunc(rv.Elem())
		argFormat = fn.argFormat()
		retFormat = fn.retFormat()
	)

	// resolve function name
	if funcName == "" {
		funcName = fn.Type().Name()
	}
	if funcName == "" {
		return errors.New("missing function name")
	}

	rv.Elem().Set(reflect.MakeFunc(fn.Type(), func(argVals []reflect.Value) (retVals []reflect.Value) {
		retVals = fn.retTemplate()

		var (
			err     error
			errType = reflect.TypeOf((*error)(nil)).Elem()
			errPtr  = reflect.ValueOf(&err)
		)

		// if last return value is an error use it to also return RPC errors
		if len(retVals) > 0 {
			lastVal := retVals[len(retVals)-1]
			if lastVal.Type().ConvertibleTo(errType) {
				errPtr = lastVal.Addr()
			}
		}

		// encode function name
		if err := cslq.Encode(rw, "[c]c", funcName); err != nil {
			errPtr.Elem().Set(reflect.ValueOf(err))
			return
		}

		// encode arguments
		if err := argFormat.Encode(rw, valuesToInterfaces(argVals)...); err != nil {
			errPtr.Elem().Set(reflect.ValueOf(err))
			return
		}

		// decode the response code
		var responseCode int
		if err := cslq.Decode(rw, "c", &responseCode); err != nil {
			errPtr.Elem().Set(reflect.ValueOf(err))
			return
		}

		if responseCode != ResponseOK {
			errPtr.Elem().Set(reflect.ValueOf(errors.New("RPC call rejected")))
			return
		}

		// decode returned values
		if err := retFormat.Decode(rw, valuesToPointers(retVals)...); err != nil {
			errPtr.Elem().Set(reflect.ValueOf(err))
			return
		}

		return
	}))

	return nil
}

func ImportStruct(rw io.ReadWriter, v interface{}) error {
	var rv = reflect.ValueOf(v)

	if (rv.Kind() != reflect.Ptr) || (rv.Elem().Kind() != reflect.Struct) {
		return errors.New("not a struct ptr")
	}

	rv = rv.Elem()

	for i := 0; i < rv.NumField(); i++ {
		if rv.Type().Field(i).IsExported() {
			if err := ImportFuncAs(rw, rv.Field(i).Addr().Interface(), rv.Type().Field(i).Name); err != nil {
				return err
			}
		}

	}

	return nil
}

func ImportFunc(rw io.ReadWriter, funcPtr interface{}) error {
	return ImportFuncAs(rw, funcPtr, "")
}
