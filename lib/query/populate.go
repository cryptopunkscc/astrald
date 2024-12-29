package query

import (
	"encoding"
	"encoding/base64"
	"errors"
	"fmt"
	"reflect"
	"strconv"
)

func Populate(m map[string]string, s any) error {
	v := reflect.ValueOf(s)

	if v.Kind() == reflect.Ptr && v.Elem().Kind() == reflect.Ptr {
		if v.Elem().IsZero() {
			v.Elem().Set(reflect.New(v.Type().Elem().Elem()))
		}
		v = v.Elem()
	}

	if v.Kind() != reflect.Ptr || v.IsZero() || v.Elem().Kind() != reflect.Struct {
		return errors.New("target must be a pointer to a struct")
	}

	v = v.Elem()

	for i := 0; i < v.NumField(); i++ {
		fv := v.Field(i)
		ft := v.Type().Field(i)

		var tags map[string]string
		lookup, ok := ft.Tag.Lookup(queryTag)
		tags = splitTag(lookup)

		if _, skip := tags["skip"]; skip {
			continue
		}

		name := toSnakeCase(ft.Name)
		if n, ok := tags["key"]; ok {
			name = n
		}
		mv, ok := m[name]
		if !ok {
			if _, optional := tags["optional"]; optional {
				continue
			}
			return fmt.Errorf("required field %s not found in the map", name)
		}

		if fv.Kind() == reflect.Ptr && fv.IsZero() {
			fv.Set(reflect.New(fv.Type().Elem()))
		}

		if fv.CanInterface() {
			if u, ok := fv.Interface().(encoding.TextUnmarshaler); ok {
				err := u.UnmarshalText([]byte(mv))
				if err != nil {
					return err
				}
				continue
			}
		}

		if fv.CanAddr() {
			fva := fv.Addr()
			if u, ok := fva.Interface().(encoding.TextUnmarshaler); ok {
				err := u.UnmarshalText([]byte(mv))
				if err != nil {
					return err
				}
				continue
			}
		}

		if fv.Kind() == reflect.Ptr {
			fv = fv.Elem()
		}

		switch fv.Kind() {
		case reflect.String:
			fv.SetString(mv)

		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			n, err := strconv.ParseInt(mv, 10, 64)
			if err != nil {
				return err
			}
			fv.SetInt(n)

		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			n, err := strconv.ParseUint(mv, 10, 64)
			if err != nil {
				return err
			}
			fv.SetUint(n)

		case reflect.Float32, reflect.Float64:
			n, err := strconv.ParseFloat(mv, 64)
			if err != nil {
				return err
			}
			fv.SetFloat(n)

		case reflect.Bool:
			if mv == "" {
				fv.SetBool(true)
				continue
			}
			n, err := strconv.ParseBool(mv)
			if err != nil {
				return err
			}
			fv.SetBool(n)

		case reflect.Slice:
			if fv.Type().Elem().Kind() == reflect.Uint8 {
				b, err := base64.StdEncoding.DecodeString(mv)
				if err != nil {
					return err
				}
				fv.Set(reflect.ValueOf(b))
			}

		default:
			return fmt.Errorf("field %s is not a supported type", ft.Name)
		}
	}

	return nil
}
