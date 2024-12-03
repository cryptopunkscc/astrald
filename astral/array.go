package astral

import (
	"io"
	"reflect"
)

// Array8 returns an untyped pseudo-object that wraps an array. Its WriteTo/ReadFrom methods invoke the respective
// method on each element of the array. Array's length is encoded as an 8-bit prefix.
func Array8[T ObjectTyper](a *[]T) Object {
	return &arr[T]{a: a, bits: 8}
}

// Array16 returns an untyped pseudo-object that wraps an array. Its WriteTo/ReadFrom methods invoke the respective
// method on each element of the array. Array's length is encoded as a 16-bit prefix.
func Array16[T ObjectTyper](a *[]T) Object {
	return &arr[T]{a: a, bits: 16}
}

// Array32 returns an untyped pseudo-object that wraps an array. Its WriteTo/ReadFrom methods invoke the respective
// method on each element of the array. Array's length is encoded as a 32-bit prefix.
func Array32[T ObjectTyper](a *[]T) Object {
	return &arr[T]{a: a, bits: 32}
}

// Array64 returns an untyped pseudo-object that wraps an array. Its WriteTo/ReadFrom methods invoke the respective
// method on each element of the array. Array's length is encoded as a 64-bit prefix.
func Array64[T ObjectTyper](a *[]T) Object {
	return &arr[T]{a: a, bits: 64}
}

var _ Object = &arr[Object]{}

type arr[T ObjectTyper] struct {
	a    *[]T
	bits int
}

type ObjectTyper interface {
	ObjectType() string
}

func (arr[T]) ObjectType() string {
	return ""
}

func (a arr[T]) WriteTo(w io.Writer) (n int64, err error) {
	v := *a.a

	n, err = a.writeLen(w, len(v))
	if err != nil {
		return
	}

	var m int64
	for _, v := range v {
		v, ok := any(v).(io.WriterTo)
		if !ok {
			panic("not a io.WriterTo")
		}
		m, err = v.WriteTo(w)
		n += m
		if err != nil {
			return
		}
	}

	return
}

func (a *arr[T]) ReadFrom(r io.Reader) (n int64, err error) {
	var l int

	n, err = a.readLen(r, &l)
	if err != nil {
		return
	}

	v := make([]T, l)
	var m int64
	for i := 0; i < int(l); i++ {
		var e T
		if reflect.TypeOf(e).Kind() == reflect.Pointer {
			e = reflect.New(reflect.TypeOf(e).Elem()).Interface().(T)
		}

		var rf io.ReaderFrom
		var ok bool

		if rf, ok = any(e).(io.ReaderFrom); !ok {
			if rf, ok = any(&e).(io.ReaderFrom); !ok {
				panic("not a io.ReaderFrom")
			}
		}

		m, err = rf.ReadFrom(r)
		n += m
		if err != nil {
			return
		}
		v[i] = e
	}
	*a.a = v

	return
}

func (a arr[T]) writeLen(w io.Writer, l int) (n int64, err error) {
	if l > (1<<a.bits)-1 {
		panic("array too long for the bit width")
	}
	switch a.bits {
	case 8:
		n, err = Uint8(l).WriteTo(w)
	case 16:
		n, err = Uint16(l).WriteTo(w)
	case 32:
		n, err = Uint32(l).WriteTo(w)
	case 64:
		n, err = Uint64(l).WriteTo(w)
	default:
		panic("unsupported bit width")
	}

	return
}

func (a arr[T]) readLen(r io.Reader, l *int) (n int64, err error) {
	switch a.bits {
	case 8:
		var k Uint8
		n, err = k.ReadFrom(r)
		*l = int(k)

	case 16:
		var k Uint16
		n, err = k.ReadFrom(r)
		*l = int(k)

	case 32:
		var k Uint32
		n, err = k.ReadFrom(r)
		*l = int(k)

	case 64:
		var k Uint64
		n, err = k.ReadFrom(r)
		*l = int(k)

	default:
		panic("unsupported bit width")
	}

	return
}
