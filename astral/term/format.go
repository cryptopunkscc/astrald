package term

import (
	"bytes"
	"reflect"
	"strconv"
	"time"

	"github.com/cryptopunkscc/astrald/astral"
)

// Format replaces %v occurances in fmt with respective values
func Format(fmt string, v ...interface{}) (list []astral.Object) {
	var lastIndex int
	for i := 0; i < len(fmt) && len(v) > 0; i++ {
		if fmt[i] == '%' && i+1 < len(fmt) && fmt[i+1] == 'v' {
			// found a %v

			var s = astral.String(fmt[lastIndex:i])
			if len(s) > 0 {
				list = append(list, &SetColor{"default"}, &s)
			}

			var arg any
			arg, v = v[0], v[1:]

			list = append(list, Objectify(arg))

			lastIndex = i + 2
			i++ // skip the next 'v'
		}
	}

	// flush the rest
	if lastIndex < len(fmt) {
		var s = astral.String(fmt[lastIndex:])
		list = append(list, &s)
	}

	return
}

func Objectify(v any) astral.Object {
	if v == nil {
		return &astral.Nil{}
	}

	// check if v is already an astral.Object
	if o, ok := v.(astral.Object); ok {
		return o
	}

	// check if a pointer to v is an Object
	objectType := reflect.TypeOf((*astral.Object)(nil)).Elem()
	if reflect.PtrTo(reflect.TypeOf(v)).Implements(objectType) {
		val := reflect.New(reflect.TypeOf(v))
		val.Elem().Set(reflect.ValueOf(v))
		return val.Interface().(astral.Object)
	}

	// convert basic and common types to their astral equivalewnts
	switch v := v.(type) {
	case PrinterTo:
		var s = astral.String(Capture(v, &DefaultTypeMap, false))
		return &s

	case error:
		var s = astral.NewError(v.Error())
		return s

	case bool:
		var b = astral.Bool(v)
		return &b

	case string:
		var s = astral.String(v)
		return &s

	case float32:
		var s = astral.String(strconv.FormatFloat(float64(v), 'f', -1, 32))
		return &s

	case float64:
		var s = astral.String(strconv.FormatFloat(v, 'f', -1, 64))
		return &s

	case uint:
		var s = astral.String(strconv.FormatUint(uint64(v), 10))
		return &s

	case uint8:
		var s = astral.String(strconv.FormatUint(uint64(v), 10))
		return &s

	case uint16:
		var s = astral.String(strconv.FormatUint(uint64(v), 10))
		return &s

	case uint32:
		var s = astral.String(strconv.FormatUint(uint64(v), 10))
		return &s

	case uint64:
		var s = astral.String(strconv.FormatUint(v, 10))
		return &s

	case int:
		var s = astral.String(strconv.FormatInt(int64(v), 10))
		return &s

	case int8:
		var s = astral.String(strconv.FormatInt(int64(v), 10))
		return &s

	case int16:
		var s = astral.String(strconv.FormatInt(int64(v), 10))
		return &s

	case int32:
		var s = astral.String(strconv.FormatInt(int64(v), 10))
		return &s

	case int64:
		var s = astral.String(strconv.FormatInt(v, 10))
		return &s

	case time.Time:
		t := astral.Time(v)
		return &t

	case time.Duration:
		d := astral.Duration(v)
		return &d

	default:
		var s = astral.String(Stringify(v))
		return &s
	}
}

func Printf(w Printer, fmt string, v ...interface{}) error {
	list := Format(fmt, v...)

	w.Print(list...)

	return nil
}

func Capture(pto PrinterTo, m Translator, mono bool) []byte {
	var buf = &bytes.Buffer{}
	var p = NewBasicPrinter(buf, m)
	p.Mono = mono
	pto.PrintTo(p)
	return buf.Bytes()
}

func Render(object astral.Object, m Translator, mono bool) []byte {
	var buf = &bytes.Buffer{}
	var p = NewBasicPrinter(buf, m)
	p.Mono = mono
	p.Print(object)
	return buf.Bytes()
}
