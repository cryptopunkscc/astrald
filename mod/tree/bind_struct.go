package tree

import (
	"errors"
	"fmt"
	"reflect"

	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/log"
)

func BindStruct(ctx *astral.Context, s any, node Node) error {
	var v = reflect.ValueOf(s)

	if v.Kind() != reflect.Ptr {
		return errors.New("s must be a pointer to a struct")
	}
	v = v.Elem()
	if v.Kind() != reflect.Struct {
		return errors.New("s must be a pointer to a struct")
	}

	for i := range v.NumField() {
		field := v.Field(i)
		fieldType := v.Type().Field(i)

		if !field.CanInterface() {
			// skip unexported fields
			continue
		}

		if field.Kind() == reflect.Ptr {
			if field.IsNil() {
				field.Set(reflect.New(field.Type().Elem()))
			}
			field = field.Elem()
		}

		if field.Kind() != reflect.Struct {
			continue
		}

		keyName := log.ToSnakeCase(fieldType.Name)

		tag := parseTag(fieldType.Tag.Get("tree"))
		if tag.path != "" {
			keyName = tag.path
		}

		bind, found := findBindMethod(field)
		if found {
			fieldNode, err := Query(ctx, node, keyName, true)
			if err != nil {
				break
			}

			ret := bind.Call([]reflect.Value{reflect.ValueOf(ctx), reflect.ValueOf(fieldNode)})

			if !ret[0].IsNil() {
				err := ret[0].Interface().(error)
				return fmt.Errorf("failed to bind field %s to key %s: %w", fieldType.Name, keyName, err)
			}
		}
	}

	return nil
}

func findBindMethod(field reflect.Value) (reflect.Value, bool) {
	if field.Kind() == reflect.Struct {
		field = field.Addr()
	}

	for j := range field.NumMethod() {
		if field.Type().Method(j).Name == "Bind" {
			bind := field.Method(j)
			return bind, true
		}
	}

	return reflect.Value{}, false
}

func parseTag(s string) tag {
	return tag{path: s}
}

type tag struct {
	path string
}
