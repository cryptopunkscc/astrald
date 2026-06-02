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

	// why: runtime Blueprints are stored under their own Type and must materialize as a
	// RuntimeObject, not as a fresh *Blueprint. The compile-time prototype itself lives under
	// "astral.blueprint" and still goes through the reflect.New path below.
	if blueprint, ok := p.(*Blueprint); ok && blueprint.Type.String() == typeName {
		return blueprint.GetRuntimeObject()
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

// RegisterBlueprint registers a runtime Blueprint with the default Blueprints and returns its ObjectID.
func RegisterBlueprint(bp *Blueprint) (*ObjectID, error) {
	return defaultBlueprints.RegisterBlueprint(bp)
}

// GetBlueprint returns the runtime *Blueprint registered under typeName, or nil if none is registered
// (compile-time prototypes are not returned).
func GetBlueprint(typeName string) *Blueprint {
	return defaultBlueprints.GetBlueprint(typeName)
}

// RegisterBlueprint stores a runtime Blueprint after structural validation. The Blueprint's Type
// must not collide with any compile-time prototype or previously registered Blueprint. Returns the
// content-addressed ObjectID of the Blueprint.
func (bp *Blueprints) RegisterBlueprint(b *Blueprint) (*ObjectID, error) {
	if err := validateBlueprint(b); err != nil {
		return nil, err
	}

	typeName := b.Type.String()
	if bp.has(typeName) {
		return nil, fmt.Errorf("%w: %s", ErrAlreadyRegistered, typeName)
	}

	_, ok := bp.Blueprints.Set(typeName, b)
	if !ok {
		// note: raced with another caller registering the same type
		return nil, fmt.Errorf("%w: %s", ErrAlreadyRegistered, typeName)
	}

	return ResolveObjectID(b)
}

// GetBlueprint returns the runtime Blueprint registered under typeName, or nil if the entry is
// absent or is a compile-time prototype (compile-time prototypes are stored under
// "astral.blueprint", never under their own runtime Type).
func (bp *Blueprints) GetBlueprint(typeName string) *Blueprint {
	o, ok := bp.Blueprints.Get(typeName)
	if !ok {
		if bp.Parent != nil {
			return bp.Parent.GetBlueprint(typeName)
		}
		return nil
	}
	b, ok := o.(*Blueprint)
	if !ok {
		return nil
	}
	// why: distinguishes the compile-time prototype (stored under "astral.blueprint") from a
	// runtime Blueprint (stored under its own Type).
	if b.Type.String() != typeName {
		return nil
	}
	return b
}

// has reports whether typeName collides with any compile-time prototype or runtime Blueprint in
// the chain.
func (bp *Blueprints) has(typeName string) bool {
	if _, ok := bp.Blueprints.Get(typeName); ok {
		return true
	}
	if bp.Parent != nil {
		return bp.Parent.has(typeName)
	}
	return false
}
