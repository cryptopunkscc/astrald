package astral

import (
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
		ro, err := blueprint.GetRuntimeObject()
		if err != nil {
			// why: validation failed (e.g. post-registration mutation produced a duplicate
			// Field.Name). Return interface-nil so the decode path surfaces the absence the
			// same way it does for unregistered types.
			return nil
		}
		return ro
	}

	var v = reflect.ValueOf(p)
	var c = reflect.New(v.Elem().Type())

	return c.Interface().(Object)
}

// Add adds new object prototypes. Returns on the first failure, so the inputs preceding the
// error are registered and the inputs following it are not — bounded partial state instead of
// scattered partial state. Pre-validates all inputs for empty type before any insertion to
// reduce TOCTOU surprises.
func (bp *Blueprints) Add(object ...Object) error {
	for _, o := range object {
		if len(o.ObjectType()) == 0 {
			return fmt.Errorf("object type is empty for %s", reflect.TypeOf(o))
		}
	}
	for _, o := range object {
		_, ok := bp.Blueprints.Set(o.ObjectType(), o)
		if !ok {
			return fmt.Errorf("blueprint for %s already added", o.ObjectType())
		}
	}
	return nil
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
//
// The caller must not mutate b (Type, Fields, or any Field.Spec) after this call returns. The
// registry stores the pointer as-is; subsequent mutations propagate to every RuntimeObject built
// from the registered Blueprint and orphan the returned ObjectID from the served schema.
//
// todo: peers are assumed to register identical Blueprint content per Type; schema divergence
// is not detected on the wire and decode silently misreads. Out of scope for now.
//
// todo: parent-chain race — insignificant for v1; production paths target
// DefaultBlueprints (no Parent) where sig.Map.Set is already race-safe.
func (bp *Blueprints) RegisterBlueprint(b *Blueprint) (*ObjectID, error) {
	if err := validateBlueprint(b); err != nil {
		return nil, err
	}

	typeName := b.Type.String()
	if bp.has(typeName) {
		return nil, fmt.Errorf("%w: %s", ErrAlreadyRegistered, typeName)
	}

	if err := bp.validateReferences(b); err != nil {
		return nil, err
	}

	// todo: think about copying blueprint
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

// validateReferences checks that every type a Blueprint's Fields point at is already
// registered in the chain. Forbids dangling refs at the registration edge — peers can't
// squat a name whose decode would then fail with ErrBlueprintNotFound. Also forbids
// mutual recursion via peer registration by design; the peer knows its type graph and
// must register prerequisites first.
func (bp *Blueprints) validateReferences(b *Blueprint) error {
	for _, f := range b.Fields {
		ref := referencedType(f.Spec)
		if ref == "" || bp.has(ref) {
			continue
		}
		return fmt.Errorf("%w: field %q references unregistered type %s",
			ErrBlueprintInvalid, f.Name, ref)
	}
	return nil
}

// referencedType returns the single external type name a Spec depends on, or "" if the
// Spec is open (heterogeneous container, ObjectSpec) or self-contained (PrimitiveSpec —
// already allowlist-checked in validateSpec).
func referencedType(spec Object) string {
	switch s := spec.(type) {
	case *RefSpec:
		return s.Type.String()
	case *PtrSpec:
		return s.Type.String()
	case *SliceSpec:
		return s.Type.String()
	case *ArraySpec:
		return s.Type.String()
	case *MapSpec:
		return s.ValueType.String()
	}
	return ""
}
