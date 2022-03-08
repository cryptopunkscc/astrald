package rpc

import (
	"github.com/cryptopunkscc/astrald/cslq"
	"reflect"
)

type funcWrapper struct {
	reflect.Value
	prefixArgs []reflect.Value
}

func wrapFuncWithPrefixArgs(value reflect.Value, prefixArgs ...interface{}) *funcWrapper {
	return &funcWrapper{Value: value, prefixArgs: interfacesToValues(prefixArgs)}
}

func wrapFunc(value reflect.Value) *funcWrapper {
	return &funcWrapper{Value: value}
}

func (a *funcWrapper) argFormat() (f cslq.Format) {
	for i := len(a.prefixArgs); i < a.Type().NumIn(); i++ {
		f = append(f, typeToOp(a.Type().In(i)))
	}
	return
}

func (a *funcWrapper) retFormat() (f cslq.Format) {
	for i := 0; i < a.Type().NumOut(); i++ {
		f = append(f, typeToOp(a.Type().Out(i)))
	}
	return
}

func (a *funcWrapper) argTemplate() (v []reflect.Value) {
	for i := len(a.prefixArgs); i < a.Type().NumIn(); i++ {
		in := a.Type().In(i)
		v = append(v, reflect.New(in).Elem())
	}
	return
}

func (a *funcWrapper) retTemplate() (v []reflect.Value) {
	for i := 0; i < a.Type().NumOut(); i++ {
		out := a.Type().Out(i)
		v = append(v, reflect.New(out).Elem())
	}
	return
}

func (a *funcWrapper) Call(in []reflect.Value) []reflect.Value {
	return a.Value.Call(append(append([]reflect.Value{}, a.prefixArgs...), in...))
}
