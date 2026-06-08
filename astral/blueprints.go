package astral

import (
	"errors"
	"fmt"
	"reflect"
	"sort"

	"github.com/cryptopunkscc/astrald/sig"
)

// Blueprints holds prototypes of astral objects. Compile-time prototypes, runtime
// Blueprints, and runtime BlueprintAliases all share one map keyed by Type name —
// each name maps to exactly one entry. Aliases for compile-time prototypes are derived
// on demand via the Aliasable interface; no parallel storage is needed.
type Blueprints struct {
	Blueprints sig.Map[string, Object]
	Parent     *Blueprints
}

// registryTodos enumerates limitations that apply equally to RegisterBlueprint and
// RegisterAlias. Kept as a single doc anchor so the comments above each method don't
// drift apart.
//
//   - schema divergence between peers under the same Type is not detected on the wire;
//     a second peer's incompatible schema returns ErrAlreadyRegistered with no detail.
//   - parent-chain race — `has()` walks the chain but `Set()` is local-only;
//     insignificant for v1 since production targets DefaultBlueprints.
//   - duplicate Register of an identical schema fails instead of being idempotent.
//   - register is not identity-gated; any peer can squat a name.
var _ = "registryTodos"

var defaultBlueprints = &Blueprints{}

func DefaultBlueprints() *Blueprints {
	return defaultBlueprints
}

// NewBlueprints returns a new Blueprints. parent is consulted by New when a type is not found locally.
func NewBlueprints(parent *Blueprints) *Blueprints {
	return &Blueprints{
		Parent: parent,
	}
}

// New returns a zero-value object of the specified type or nil if no blueprint is found.
func New(typeName string) Object {
	return defaultBlueprints.New(typeName)
}

// Add registers object prototypes with the default Blueprints.
func Add(object ...Object) error {
	return defaultBlueprints.Add(object...)
}

// GetAlias returns the runtime *BlueprintAlias for typeName, or nil.
func GetAlias(typeName string) *BlueprintAlias {
	return defaultBlueprints.GetAlias(typeName)
}

// New returns a zero-value object of the specified type or nil if no blueprint is found.
//
// The stored entry under typeName disambiguates: a *Blueprint registered under its own Type
// materializes as *RuntimeObject; a *BlueprintAlias registered under its own Type
// materializes as *RuntimeAlias; anything else is a compile-time prototype handed back via
// reflect.New (giving the originating peer the typed Go value rather than a wire carrier).
func (bp *Blueprints) New(typeName string) Object {
	p, ok := bp.Blueprints.Get(typeName)
	if !ok {
		if bp.Parent != nil {
			return bp.Parent.New(typeName)
		}
		return nil
	}

	// why: runtime Blueprint/BlueprintAlias stored under their own Type materialize as
	// the matching runtime carrier; compile-time prototypes live under their prototype
	// name with empty Type and fall through to reflect.New. A GetRuntime* failure means
	// post-registration mutation broke validation — return interface-nil so decode
	// surfaces the absence the same way as an unregistered type.
	switch v := p.(type) {
	case *Blueprint:
		if v.Type.String() == typeName {
			ro, err := v.GetRuntimeObject()
			if err != nil {
				return nil
			}
			return ro
		}
	case *BlueprintAlias:
		if v.Type.String() == typeName {
			ra, err := v.GetRuntimeAlias()
			if err != nil {
				return nil
			}
			return ra
		}
	}

	return reflect.New(reflect.ValueOf(p).Elem().Type()).Interface().(Object)
}

// Add registers object prototypes. Pre-validates empty types before any insertion.
// Returns on first failure; inputs before the error are registered (bounded partial state).
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

// Types returns names of all registered object types.
func (bp *Blueprints) Types() (names []string) {
	if bp.Parent != nil {
		names = bp.Parent.Types()
	}
	return append(names, bp.Blueprints.Keys()...)
}

// AllBlueprints returns every registered prototype as a *Blueprint, in
// dependency order. Compile-time entries are derived via BlueprintOf; runtime
// *Blueprints pass through. Per-entry derivation failures are aggregated
// into err; the returned slice contains the successful entries.
func (bp *Blueprints) AllBlueprints() ([]*Blueprint, error) {
	var out []*Blueprint
	var errs []error
	if bp.Parent != nil {
		parent, perr := bp.Parent.AllBlueprints()
		out = append(out, parent...)
		if perr != nil {
			errs = append(errs, perr)
		}
	}

	var proto []*Blueprint
	var runtime []*Blueprint
	for name, obj := range bp.Blueprints.Clone() {
		if b, ok := obj.(*Blueprint); ok && b.Type.String() == name {
			runtime = append(runtime, b)
			continue
		}
		// why: aliases are returned by AllAliases; skip both stored aliases and
		// compile-time prototypes that satisfy Aliasable so BlueprintOf isn't asked
		// to derive a struct schema for a primitive carrier.
		if a, ok := obj.(*BlueprintAlias); ok && a.Type.String() == name {
			continue
		}
		if _, ok := obj.(Aliasable); ok {
			continue
		}

		derived, err := BlueprintOf(obj)
		if err != nil {
			errs = append(errs, fmt.Errorf("blueprint %s: %w", name, err))
			continue
		}
		proto = append(proto, derived)
	}

	sort.Slice(proto, func(i, j int) bool {
		return proto[i].Type.String() < proto[j].Type.String()
	})
	out = append(out, proto...)

	for _, name := range orderBlueprintsByReference(runtime) {
		for _, b := range runtime {
			if b.Type.String() == name {
				out = append(out, b)
				break
			}
		}
	}

	if len(errs) == 0 {
		return out, nil
	}
	return out, errors.Join(errs...)
}

// OrderedBlueprints returns all registered type names in dependency order. Walks
// the parent chain. At each level: non-alias compile-time prototypes first
// (alpha-sorted), then aliases (alpha-sorted, leaves with no internal refs),
// then runtime Blueprints topo-sorted by referencedType.
//
// Aliases precede runtime Blueprints so that a Blueprint's RefSpec to an alias
// resolves on the peer when replayed in this order. An entry classifies as an
// alias when it's a stored *BlueprintAlias OR a compile-time prototype that
// satisfies Aliasable.
func (bp *Blueprints) OrderedBlueprints() []string {
	var out []string
	if bp.Parent != nil {
		out = bp.Parent.OrderedBlueprints()
	}

	var proto []string
	var aliases []string
	var runtime []*Blueprint
	for name, obj := range bp.Blueprints.Clone() {
		if b, ok := obj.(*Blueprint); ok && b.Type.String() == name {
			runtime = append(runtime, b)
			continue
		}
		if a, ok := obj.(*BlueprintAlias); ok && a.Type.String() == name {
			aliases = append(aliases, name)
			continue
		}
		if _, ok := obj.(Aliasable); ok {
			aliases = append(aliases, name)
			continue
		}
		proto = append(proto, name)
	}

	sort.Strings(proto)
	sort.Strings(aliases)
	out = append(out, proto...)
	out = append(out, aliases...)
	out = append(out, orderBlueprintsByReference(runtime)...)
	return out
}

// orderBlueprintsByReference returns blueprint names ordered (Kahn-style topological sort) so each precedes
// any that references it. References outside the input set are treated as
// satisfied. Alpha tie-break.
func orderBlueprintsByReference(bps []*Blueprint) []string {
	if len(bps) == 0 {
		return nil
	}

	inSet := make(map[string]bool, len(bps))
	for _, b := range bps {
		inSet[b.Type.String()] = true
	}

	inDeg := make(map[string]int, len(bps))
	adj := make(map[string][]string, len(bps))
	for _, b := range bps {
		name := b.Type.String()
		inDeg[name] = inDeg[name] // touch
		for _, f := range b.Fields {
			ref := referencedType(f.Spec)
			if ref == "" || !inSet[ref] || ref == name {
				continue
			}
			adj[ref] = append(adj[ref], name)
			inDeg[name]++
		}
	}

	names := make([]string, 0, len(bps))
	for _, b := range bps {
		names = append(names, b.Type.String())
	}
	sort.Strings(names)

	emitted := make(map[string]bool, len(bps))
	out := make([]string, 0, len(bps))
	for len(out) < len(names) {
		progress := false
		for _, n := range names {
			if emitted[n] || inDeg[n] != 0 {
				continue
			}
			out = append(out, n)
			emitted[n] = true
			for _, d := range adj[n] {
				inDeg[d]--
			}
			progress = true
		}
		if !progress {
			for _, n := range names {
				if !emitted[n] {
					out = append(out, n)
					emitted[n] = true
				}
			}
			break
		}
	}
	return out
}

// Register stores a runtime schema descriptor (*Blueprint or *BlueprintAlias) in the
// default Blueprints and returns its content-addressed ObjectID. The canonical entry
// point for runtime registration — typed methods on *Blueprints (RegisterBlueprint,
// RegisterAlias) remain for callers that already hold the concrete type.
func Register(o Object) (*ObjectID, error) {
	return defaultBlueprints.Register(o)
}

// Register dispatches to RegisterBlueprint or RegisterAlias based on the concrete type
// of o. Returns ErrBlueprintInvalid for anything that isn't a *Blueprint or *BlueprintAlias.
func (bp *Blueprints) Register(o Object) (*ObjectID, error) {
	switch v := o.(type) {
	case *Blueprint:
		return bp.RegisterBlueprint(v)
	case *BlueprintAlias:
		return bp.RegisterAlias(v)
	default:
		return nil, fmt.Errorf("%w: Register: want *Blueprint or *BlueprintAlias, got %T",
			ErrBlueprintInvalid, o)
	}
}

// GetBlueprint returns the runtime *Blueprint for typeName, or nil.
// Compile-time prototypes are not returned.
func GetBlueprint(typeName string) *Blueprint {
	return defaultBlueprints.GetBlueprint(typeName)
}

// RegisterBlueprint stores a runtime Blueprint after validation and returns its
// content-addressed ObjectID. Type must not collide with any compile-time prototype,
// previously registered Blueprint, or registered BlueprintAlias — all share the same
// Blueprints map.
//
// Caller must not mutate b after this call. The registry stores the pointer as-is;
// mutations propagate to every RuntimeObject and orphan the returned ObjectID.
//
// See registryTodos for limitations shared with RegisterAlias.
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

// GetBlueprint returns the runtime Blueprint for typeName, or nil. Compile-time prototypes
// live under "astral.blueprint", never under their own runtime Type, so they return nil.
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
	// why: distinguishes the compile-time prototype from a runtime Blueprint.
	if b.Type.String() != typeName {
		return nil
	}
	return b
}

// has reports whether typeName exists anywhere in the chain.
func (bp *Blueprints) has(typeName string) bool {
	if _, ok := bp.Blueprints.Get(typeName); ok {
		return true
	}
	if bp.Parent != nil {
		return bp.Parent.has(typeName)
	}
	return false
}

// validateReferences requires every referenced type to be already registered.
// Forbids dangling refs (would fail decode with ErrBlueprintNotFound) and mutual
// recursion (peers must register prerequisites first).
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

// referencedType returns the type name a Spec depends on, or "" for open specs
// (heterogeneous containers, ObjectSpec) and self-contained PrimitiveSpec.
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

// RegisterAlias stores a runtime BlueprintAlias after validation and returns its
// content-addressed ObjectID. Shares the main registry map with Blueprints and
// compile-time prototypes — any name collision returns ErrAlreadyRegistered.
//
// Caller must not mutate a after this call.
//
// See registryTodos for limitations shared with RegisterBlueprint.
func (bp *Blueprints) RegisterAlias(a *BlueprintAlias) (*ObjectID, error) {
	if err := validateAlias(a); err != nil {
		return nil, err
	}

	typeName := a.Type.String()
	if bp.has(typeName) {
		return nil, fmt.Errorf("%w: %s", ErrAlreadyRegistered, typeName)
	}

	_, ok := bp.Blueprints.Set(typeName, a)
	if !ok {
		// note: raced with another caller registering the same name.
		return nil, fmt.Errorf("%w: %s", ErrAlreadyRegistered, typeName)
	}

	return ResolveObjectID(a)
}

// GetAlias returns the runtime BlueprintAlias stored under typeName, or nil. Compile-time
// prototypes (including Aliasable ones) are not returned — use AllAliases for the derived
// form, or inspect the prototype directly via the Aliasable assertion.
func (bp *Blueprints) GetAlias(typeName string) *BlueprintAlias {
	if p, ok := bp.Blueprints.Get(typeName); ok {
		if a, ok := p.(*BlueprintAlias); ok && a.Type.String() == typeName {
			return a
		}
	}
	if bp.Parent != nil {
		return bp.Parent.GetAlias(typeName)
	}
	return nil
}

// AllSchemas returns every runtime schema descriptor (aliases + Blueprints) ordered for
// sync replay: aliases first (alpha within each parent level, leaves with no refs), then
// runtime Blueprints topo-sorted by reference. Parent-chain entries precede local ones.
// Derivation failures from AllAliases / AllBlueprints are aggregated into err; the
// returned slice contains the successful entries.
//
// Callers that only want one half can use AllAliases or AllBlueprints; AllSchemas exists
// so sync can do a single walk and produce one ordered cache.
func (bp *Blueprints) AllSchemas() ([]Object, error) {
	aliases, aerr := bp.AllAliases()
	blueprints, berr := bp.AllBlueprints()

	out := make([]Object, 0, len(aliases)+len(blueprints))
	for _, a := range aliases {
		out = append(out, a)
	}
	for _, b := range blueprints {
		out = append(out, b)
	}

	switch {
	case aerr == nil && berr == nil:
		return out, nil
	case aerr == nil:
		return out, berr
	case berr == nil:
		return out, aerr
	}
	return out, errors.Join(aerr, berr)
}

// AllAliases returns every alias known at this level: stored *BlueprintAlias entries plus
// entries derived from compile-time prototypes that implement Aliasable. Parent-chain
// entries come first; at each level aliases are alpha-sorted by Type. Derivation failures
// (Aliasable.UnderlyingPrimitive returning a value outside primitiveAllowlist) are
// aggregated into err; the returned slice contains the successful entries.
func (bp *Blueprints) AllAliases() ([]*BlueprintAlias, error) {
	var out []*BlueprintAlias
	var errs []error
	if bp.Parent != nil {
		parent, perr := bp.Parent.AllAliases()
		out = append(out, parent...)
		if perr != nil {
			errs = append(errs, perr)
		}
	}

	local := make([]*BlueprintAlias, 0)
	for name, obj := range bp.Blueprints.Clone() {
		if a, ok := obj.(*BlueprintAlias); ok && a.Type.String() == name {
			local = append(local, a)
			continue
		}
		if _, ok := obj.(Aliasable); !ok {
			continue
		}
		derived, err := AliasOf(obj)
		if err != nil {
			errs = append(errs, fmt.Errorf("alias %s: %w", name, err))
			continue
		}
		local = append(local, derived)
	}
	sort.Slice(local, func(i, j int) bool {
		return local[i].Type.String() < local[j].Type.String()
	})
	out = append(out, local...)

	if len(errs) == 0 {
		return out, nil
	}
	return out, errors.Join(errs...)
}
