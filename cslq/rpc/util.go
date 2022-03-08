package rpc

import (
	"github.com/cryptopunkscc/astrald/cslq"
	"reflect"
)

func valuesToPointers(values []reflect.Value) (ptrs []interface{}) {
	for _, i := range values {
		ptrs = append(ptrs, i.Addr().Interface())
	}
	return
}

func valuesToInterfaces(values []reflect.Value) (ifaces []interface{}) {
	for _, i := range values {
		ifaces = append(ifaces, i.Interface())
	}
	return
}

func interfacesToValues(ifaces []interface{}) (vals []reflect.Value) {
	for _, i := range ifaces {
		vals = append(vals, reflect.ValueOf(i))
	}
	return
}

func typeToOp(t reflect.Type) cslq.Op {
	switch t.Kind() {
	case reflect.Uint8, reflect.Int8:
		return cslq.OpUint8{}
	case reflect.Uint16, reflect.Int16:
		return cslq.OpUint16{}
	case reflect.Uint32, reflect.Int32:
		return cslq.OpUint32{}
	case reflect.Uint64, reflect.Uint, reflect.Int:
		return cslq.OpUint64{}
	case reflect.String:
		return cslq.OpArray{
			LenOp:  cslq.OpUint64{},
			ElemOp: cslq.OpUint8{},
		}
	case reflect.Array, reflect.Slice:
		return cslq.OpArray{
			LenOp:  cslq.OpUint64{},
			ElemOp: typeToOp(t.Elem()),
		}
	case reflect.Ptr, reflect.Struct:
		return cslq.OpStruct{}
	default:
		return cslq.OpInterface{}
	}
}

func isStructOrPtr(rv reflect.Value) bool {
	return rv.Kind() == reflect.Struct || (rv.Kind() == reflect.Ptr && rv.Elem().Kind() == reflect.Struct)
}
