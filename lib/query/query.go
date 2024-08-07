package query

import (
	"errors"
	"fmt"
	"net/url"
	"reflect"
	"strconv"
	"strings"
	"unicode"
)

const queryTag = "query"

func Parse(q string) (path string, params map[string]string) {
	var s string
	path, s = splitPathParams(q)
	params = map[string]string{}

	vals, err := url.ParseQuery(s)
	if err != nil {
		return
	}

	for k, v := range vals {
		if len(v) > 0 {
			params[k] = v[0]
		} else {
			params[k] = ""
		}
	}

	return
}

func ParseTo(q string, args any) (path string, err error) {
	path, params := Parse(q)
	err = Populate(params, args)
	return
}

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

		default:
			return fmt.Errorf("field %s is not a supported type", ft.Name)
		}
	}

	return nil
}

func toSnakeCase(str string) string {
	var result []rune
	var lastUpper bool
	for i, r := range str {
		if unicode.IsUpper(r) {
			if i > 0 && !lastUpper {
				result = append(result, '_')
			}
			result = append(result, unicode.ToLower(r))
			lastUpper = true
		} else {
			result = append(result, r)
			lastUpper = false
		}
	}
	return string(result)
}

func splitTag(tag string) (m map[string]string) {
	m = make(map[string]string)

	s := strings.Split(tag, ";")
	for _, v := range s {
		p := strings.SplitN(v, ":", 2)
		if len(p) < 2 {
			m[p[0]] = ""
		} else {
			m[p[0]] = p[1]
		}
	}

	return m
}

func splitPathParams(query string) (path, params string) {
	if i := strings.IndexByte(query, '?'); i != -1 {
		return query[:i], query[i+1:]
	}
	return query, ""
}
