package cslq

import (
	"fmt"
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

func expectToken(tokens TokenReader, expect Token) error {
	next, err := tokens.Read()
	if err != nil {
		return err
	}
	nextType, expectType := reflect.TypeOf(next), reflect.TypeOf(expect)
	if nextType != expectType {
		return fmt.Errorf("expected %s, got %s", expectType, nextType)
	}
	return nil
}
