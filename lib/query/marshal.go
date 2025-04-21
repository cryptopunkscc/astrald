package query

import (
	"encoding"
	"encoding/base64"
	"errors"
	"fmt"
	"github.com/cryptopunkscc/astrald/lib/term"
	"net/url"
	"reflect"
	"strconv"
)

func Marshal(a any) (string, error) {
	if a == nil {
		return "", nil
	}

	var vals = url.Values{}

	if m, ok := a.(map[string]string); ok {
		for k, v := range m {
			vals.Set(k, v)
		}
		return vals.Encode(), nil
	}

	if m, ok := a.(Args); ok {
		for k, v := range m {
			vals.Set(k, v)
		}
		return vals.Encode(), nil
	}

	var v = reflect.ValueOf(a)
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

		if _, skip := tags["skip"]; skip {
			continue
		}

		name := term.ToSnakeCase(ft.Name)
		if n, ok := tags["key"]; ok {
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
