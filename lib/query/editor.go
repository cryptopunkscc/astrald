package query

import (
	"encoding/json"
	"errors"
	"net/url"
	"reflect"

	"github.com/cryptopunkscc/astrald/astral/log"
)

// Editor wraps a struct and provides getter/setters for the fields
type Editor struct {
	arg    reflect.Value // value of the struct (dereferenced)
	fields []*field
}

type field struct {
	name string
	*FieldEditor
}

// FieldSpec describes a field in the Editor struct
type FieldSpec struct {
	Name     string
	Type     string
	Required bool
}

// Edit wraps a struct with an Editor
func Edit(args any) (*Editor, error) {
	return edit(args, false)
}

func EditCamel(args any) (*Editor, error) {
	return edit(args, true)
}
func edit(args any, keepCase bool) (*Editor, error) {
	var v = reflect.ValueOf(args)

	// make sure arg is a pointer to a struct so we can modify it
	if v.Kind() == reflect.Ptr && v.Elem().Kind() == reflect.Ptr {
		if v.Elem().IsZero() {
			v.Elem().Set(reflect.New(v.Type().Elem().Elem()))
		}
		v = v.Elem()
	}
	if v.Kind() != reflect.Ptr || v.IsZero() || v.Elem().Kind() != reflect.Struct {
		return nil, errors.New("arg must be a pointer to a struct")
	}
	v = v.Elem()

	view := &Editor{
		arg: v,
	}

	for i := 0; i < v.NumField(); i++ {
		fv := v.Field(i)
		ft := v.Type().Field(i)

		if !fv.CanInterface() {
			continue
		}

		editor := newFieldEditor(fv, ft)

		// don't include fields with the skip tag
		if editor.Tag().Skip {
			continue
		}

		name := ft.Name

		if !keepCase {
			name = log.ToSnakeCase(name)
		}

		if editor.Tag().Key != "" {
			name = editor.Tag().Key
		}

		view.fields = append(view.fields, &field{
			name:        name,
			FieldEditor: editor,
		})
	}

	return view, nil
}

func (args *Editor) Set(name string, value string) error {
	field, err := args.Field(name)
	if err != nil {
		return err
	}

	return field.Set(value)
}

func (args *Editor) Get(name string) (string, error) {
	field, err := args.Field(name)
	if err != nil {
		return "", err
	}

	return field.Get(), nil
}

// Spec returns a map containing specs for every argument
func (args *Editor) Spec() (vals []FieldSpec) {
	for _, editor := range args.fields {
		if editor.ObjectType() == "" {
			continue
		}
		vals = append(vals, FieldSpec{
			Name:     editor.name,
			Type:     editor.ObjectType(),
			Required: !editor.Tag().Optional,
		})
	}
	return
}

func (args *Editor) Field(name string) (*FieldEditor, error) {
	for _, field := range args.fields {
		if field.name == name {
			return field.FieldEditor, nil
		}
	}

	return nil, errors.New("field not found")
}

func (args *Editor) SetMany(vals map[string]string) error {
	for key, value := range vals {
		err := args.Set(key, value)
		if err != nil {
			return err
		}
	}
	return nil
}

// json (passthrough)

func (args *Editor) MarshalJSON() ([]byte, error) {
	return json.Marshal(args.arg.Interface())
}

func (args *Editor) UnmarshalJSON(bytes []byte) error {
	return json.Unmarshal(bytes, args.arg.Addr().Interface())
}

// query

func (args *Editor) MarshalQuery() ([]byte, error) {
	var vals = url.Values{}

	for _, field := range args.fields {
		text, err := field.MarshalText()
		if err != nil {
			return nil, err
		}
		if text == nil {
			continue
		}

		vals.Set(field.name, string(text))
	}

	return []byte(vals.Encode()), nil
}

func (args *Editor) UnmarshalQuery(text []byte) error {
	vals, err := url.ParseQuery(string(text))
	if err != nil {
		return err
	}

	for key, value := range vals {
		err = args.Set(key, value[0])
		if err != nil {
			return err
		}
	}
	return nil
}

func (args *Editor) String() string {
	text, _ := args.MarshalQuery()
	return string(text)
}
