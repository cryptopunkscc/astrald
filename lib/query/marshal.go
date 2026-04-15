package query

import (
	"encoding"
	"errors"
	"fmt"
	"net/url"
	"reflect"
)

const queryTag = "query"

// Marshal marshals params provided in the argument to a string. params can be a string, a map[string]string,
// a map[string]any or a struct
func Marshal(params any) (string, error) {
	if params == nil {
		return "", nil
	}

	var vals = url.Values{}

	switch a := params.(type) {
	case string:
		return a, nil

	case map[string]string:
		for k, v := range a {
			vals.Set(k, v)
		}
		return vals.Encode(), nil

	case map[string]any:
		for k, v := range a {
			s := ""

			switch v := v.(type) {
			case string:
				s = v

			case encoding.TextMarshaler:
				text, err := v.MarshalText()
				if err != nil {
					return "", err
				}
				s = string(text)

			case fmt.Stringer:
				s = v.String()

			default:
				s = fmt.Sprintf("%v", v)
			}

			vals.Set(k, s)
		}

		return vals.Encode(), nil

	case Args:
		return Marshal(map[string]any(a))
	}

	var v = reflect.ValueOf(params)
	if v.Kind() == reflect.Ptr && v.Elem().Kind() == reflect.Struct {
		str, err := EditValue(v).MarshalQuery()
		if err != nil {
			return "", err
		}
		return string(str), nil
	}

	if v.Kind() == reflect.Struct {
		if v.CanAddr() {
			str, err := EditValue(v.Addr()).MarshalQuery()
			if err != nil {
				return "", err
			}
			return string(str), nil
		}

		pv := reflect.New(v.Type())
		pv.Elem().Set(v)

		str, err := EditValue(pv).MarshalQuery()
		if err != nil {
			return "", err
		}
		return string(str), nil
	}

	return "", errors.New("unsupported type: " + v.Kind().String())
}
