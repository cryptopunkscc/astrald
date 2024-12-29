package term

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/sig"
	"reflect"
)

type Translator interface {
	Translate(astral.Object) astral.Object
}

type TranslateFunc func(astral.Object) astral.Object

type TypeMap struct {
	p Translator
	t sig.Map[string, Translator]
}

func NewTypeMap(parent Translator) *TypeMap {
	return &TypeMap{p: parent}
}

func (m *TypeMap) SetTranslator(typeName string, t Translator) error {
	m.t.Set(typeName, t)
	return nil
}

func (m *TypeMap) SetTranslateFunc(typeName string, t TranslateFunc) error {
	m.t.Replace(typeName, adapter{t})
	return nil
}

func (m *TypeMap) Translate(object astral.Object) astral.Object {
	for i := maxTranslateDepth; i > 0; i-- {
		if t, found := m.t.Get(object.ObjectType()); found {
			object = t.Translate(object)
			continue
		}
		if m.p != nil {
			o := m.p.Translate(object)
			if o != object {
				// if the parent translated the object, keep trying
				object = o
				continue
			}
		}

		break
	}
	return object
}

type adapter struct {
	TranslateFunc
}

func (a adapter) Translate(object astral.Object) astral.Object { return a.TranslateFunc(object) }

func SetTranslateFunc[T astral.Object](fn func(o T) astral.Object) {
	var t T

	if reflect.TypeOf(t).Kind() == reflect.Ptr {
		t = reflect.New(reflect.TypeOf(t).Elem()).Interface().(T)
	}

	var typeName = t.ObjectType()
	DefaultTypeMap.SetTranslateFunc(typeName, func(object astral.Object) astral.Object {
		var ok bool
		t, ok = object.(T)
		if !ok {
			return object
		}
		return fn(t)
	})
}
