package astral

import (
	"bytes"
	"io"
	"reflect"
)

var _ Object = &Slice[Object]{}

type Slice[T ObjectTyper] struct {
	Elem     *[]T
	LenBits  int
	ElemBits int
}

type ObjectTyper interface {
	ObjectType() string
}

func WrapSlice[T ObjectTyper](elem *[]T, lenBits int, elemBits int) *Slice[T] {
	return &Slice[T]{
		Elem:     elem,
		LenBits:  lenBits,
		ElemBits: elemBits,
	}
}

func (Slice[T]) ObjectType() string {
	return ""
}

func (a Slice[T]) WriteTo(w io.Writer) (n int64, err error) {
	v := *a.Elem

	n, err = writeInt(w, len(v), a.LenBits)
	if err != nil {
		return
	}

	var m int64
	for _, v := range v {
		wto, ok := any(v).(io.WriterTo)
		if !ok {
			panic("not an io.WriterTo")
		}

		var buf = &bytes.Buffer{}

		_, err = wto.WriteTo(buf)
		if err != nil {
			return
		}

		m, err = writeInt(w, len(buf.Bytes()), a.ElemBits)
		n += m
		if err != nil {
			return
		}

		var j int
		j, err = w.Write(buf.Bytes())
		n += int64(j)
		if err != nil {
			return
		}
	}

	return
}

func (a *Slice[T]) ReadFrom(r io.Reader) (n int64, err error) {
	var l int

	n, err = loadInt(r, &l, a.LenBits)
	if err != nil {
		return
	}

	v := make([]T, l)

	var m int64
	for i := 0; i < l; i++ {
		var e T
		if reflect.TypeOf(e).Kind() == reflect.Pointer {
			e = reflect.New(reflect.TypeOf(e).Elem()).Interface().(T)
		}

		// read element length
		var el int
		m, err = loadInt(r, &el, a.ElemBits)
		n += m
		if err != nil {
			return
		}

		// read element bytes
		var buf = make([]byte, el)
		var j int
		j, err = io.ReadFull(r, buf)
		n += int64(j)
		if err != nil {
			return
		}

		var rf io.ReaderFrom
		var ok bool

		if rf, ok = any(e).(io.ReaderFrom); !ok {
			if rf, ok = any(&e).(io.ReaderFrom); !ok {
				panic("slice element is not an io.ReaderFrom")
			}
		}

		_, err = rf.ReadFrom(bytes.NewReader(buf))
		if err != nil {
			return
		}
		v[i] = e
	}
	*a.Elem = v

	return
}

func writeInt(w io.Writer, l int, bits int) (n int64, err error) {
	if l > (1<<bits)-1 {
		panic("array too long for the bit width")
	}
	switch bits {
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

func loadInt(r io.Reader, l *int, bits int) (n int64, err error) {
	switch bits {
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
