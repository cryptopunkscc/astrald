package query

import (
	"encoding"
	"encoding/base64"
	"errors"
	"fmt"
	"net/url"
	"reflect"
	"strconv"

	"github.com/cryptopunkscc/astrald/lib/term"
)

const queryTag = "query"
const skipTag = "skip"
const keyTag = "key"
const optionalTag = "optional"

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
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	if v.Kind() != reflect.Struct {
		return "", errors.New("not a struct")
	}

	for i := 0; i < v.NumField(); i++ {
		fv := v.Field(i)
		ft := v.Type().Field(i)

		var tags map[string]string
		lookup, _ := ft.Tag.Lookup(queryTag)
		tags = splitTag(lookup)

		if _, skip := tags[skipTag]; skip {
			continue
		}

		name := term.ToSnakeCase(ft.Name)
		if n, ok := tags[keyTag]; ok {
			name = n
		}

		if fv.IsZero() {
			continue
		}

		if fv.CanInterface() {
			if u, ok := fv.Interface().(encoding.TextMarshaler); ok {
				text, err := u.MarshalText()
				if err != nil {
					return "", err
				}
				vals.Set(name, string(text))
				continue
			}
		}

		if fv.CanAddr() {
			fva := fv.Addr()
			if u, ok := fva.Interface().(encoding.TextMarshaler); ok {
				text, err := u.MarshalText()
				if err != nil {
					return "", err
				}
				vals.Set(name, string(text))
				continue
			}
		}

		if fv.Kind() == reflect.Ptr {
			fv = fv.Elem()
		}

		switch fv.Kind() {
		case reflect.String:
			vals.Set(name, fv.String())

		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			vals.Set(name, strconv.FormatInt(fv.Int(), 10))

		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			vals.Set(name, strconv.FormatUint(fv.Uint(), 10))

		case reflect.Float32, reflect.Float64:
			vals.Set(name, strconv.FormatFloat(fv.Float(), 'g', -1, 64))

		case reflect.Bool:
			vals.Set(name, strconv.FormatBool(fv.Bool()))

		case reflect.Slice:
			if fv.Type().Elem().Kind() == reflect.Uint8 {
				vals.Set(name, base64.StdEncoding.EncodeToString(fv.Interface().([]byte)))
			} else {
				return "", fmt.Errorf("field %s is not a supported type", ft.Name)
			}

		default:
			return "", fmt.Errorf("field %s is not a supported type", ft.Name)
		}
	}

	return vals.Encode(), nil
}
