package cslq

import (
	"reflect"
	"strings"
)

func extractStructFields(rv reflect.Value) []interface{} {
	vars := make([]interface{}, 0)

	for i := 0; i < rv.NumField(); i++ {
		field := rv.Field(i)
		if !field.CanInterface() {
			continue
		}
		if strings.Contains(rv.Type().Field(i).Tag.Get(tagCSLQ), tagSkip) {
			continue
		}
		if field.CanAddr() {
			vars = append(vars, field.Addr().Interface())
		} else {
			vars = append(vars, field.Interface())
		}
	}

	return vars
}
