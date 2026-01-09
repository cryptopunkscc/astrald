package astral

import (
	"errors"
	"fmt"
	"reflect"

	"github.com/cryptopunkscc/astrald/sig"
)

// Blueprints is a structure that holds prototypes of astral objects.
type Blueprints struct {
	Blueprints sig.Map[string, Object]
	Parent     *Blueprints
}

var defaultBlueprints = &Blueprints{}

func DefaultBlueprints() *Blueprints {
	return defaultBlueprints
}

// NewBlueprints returns a new instance of Blueprints. If parent is not nil, it will be used by New() to look up
// prototypes if not found in this instance.
func NewBlueprints(parent *Blueprints) *Blueprints {
	return &Blueprints{
		Parent: parent,
	}
}

// New returns a zero-value object of the specified type or nil if no blueprint is found.
func New(typeName string) Object {
	return defaultBlueprints.New(typeName)
}

// Add adds the object prototypes to the default Blueprints
func Add(object ...Object) error {
	return defaultBlueprints.Add(object...)
}

// New returns a zero-value object of the specified type or nil if no blueprint is found.
func (bp *Blueprints) New(typeName string) Object {
	p, ok := bp.Blueprints.Get(typeName)
	if !ok {
		if bp.Parent != nil {
			return bp.Parent.New(typeName)
		}
		return nil
	}
	var v = reflect.ValueOf(p)
	var c = reflect.New(v.Elem().Type())

	return c.Interface().(Object)
}

// Add adds a new object prototype
func (bp *Blueprints) Add(object ...Object) error {
	var errs []error

	for _, o := range object {
		if len(o.ObjectType()) == 0 {
			errs = append(errs, fmt.Errorf("object type is empty for %s", reflect.TypeOf(o)))
			continue
		}
		_, ok := bp.Blueprints.Set(o.ObjectType(), o)
		if !ok {
			errs = append(errs, fmt.Errorf("blueprint for %s already added", o.ObjectType()))
		}
	}

	return errors.Join(errs...)
}

// Types returns type names of all registered object types
func (bp *Blueprints) Types() (names []string) {
	if bp.Parent != nil {
		names = bp.Parent.Types()
	}
	return append(names, bp.Blueprints.Keys()...)
}
