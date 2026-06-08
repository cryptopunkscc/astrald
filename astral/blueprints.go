package astral

import (
	"errors"
	"fmt"
	"reflect"
	"sort"

	"github.com/cryptopunkscc/astrald/sig"
)

// Blueprints holds prototypes of astral objects. Compile-time prototypes and runtime
// Blueprints (both struct kind and alias kind) all share one map keyed by Type name —
// each name maps to exactly one entry. Aliases for compile-time prototypes are derived
// on demand via the PrimitiveAlias interface; no parallel storage is needed.
type Blueprints struct {
	entries sig.Map[string, Object]
	Parent  *Blueprints
}

// entryKind classifies a Blueprints map entry by how it should be materialized
// or surfaced. The dispatch (runtime carrier vs. derived alias vs. struct
// prototype) is identical across New, AllBlueprints, and OrderedBlueprints;
// classify centralizes that switch.
type entryKind int

const (
	kindStructProto entryKind = iota
	kindRuntimeBP
	kindAliasProto
)

// classify reports the entryKind of obj stored under name. A *Blueprint counts as the
// runtime kind only when its Type matches the key — otherwise it's the compile-time
// prototype of Blueprint itself. Stored *Blueprint runtime entries cover both struct
// and alias kinds; their internal kind is read off the Blueprint when needed.
func classify(name string, obj Object) entryKind {
	if b, ok := obj.(*Blueprint); ok && b.Type.String() == name {
		return kindRuntimeBP
	}

	if _, ok := obj.(PrimitiveAlias); ok {
		return kindAliasProto
	}

	return kindStructProto
}

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

// New returns a zero-value object of the specified type or nil if no blueprint is found.
//
// The stored entry under typeName disambiguates: a *Blueprint registered under its own Type
// materializes as *RuntimeObject (dispatching internally on Kind for struct vs alias);
// anything else is a compile-time prototype handed back via reflect.New (giving the
// originating peer the typed Go value rather than a wire carrier).
func (bp *Blueprints) New(typeName string) Object {
	o, _ := newAt(bp, typeName, 0)
	return o
}

// newAt is the depth-aware internal entry. specZeroAt threads its construction depth here
// so a RefSpec/PtrSpec cycle materializing one runtime Blueprint after another is bounded
// by MaxBlueprintDepth rather than overflowing the Go stack. Returns the constructor's
// error when it's ErrDepthExceeded so the outer construction can surface it; other errors
// are intentionally swallowed to nil to match the documented "treat as unregistered" contract.
func newAt(bp *Blueprints, typeName string, depth int) (Object, error) {
	p, ok := bp.entries.Get(typeName)
	if !ok {
		if bp.Parent != nil {
			return newAt(bp.Parent, typeName, depth)
		}
		return nil, nil
	}

	// why: a runtime Blueprint stored under its own Type materializes as RuntimeObject;
	// compile-time prototypes live under their prototype name with empty Type and fall
	// through to reflect.New. A constructor failure means post-registration mutation
	// broke validation — return interface-nil so decode surfaces the absence the same
	// way as an unregistered type. Depth-exceeded escapes the nil-swallow so the
	// originating construction surfaces a typed error.
	if classify(typeName, p) == kindRuntimeBP {
		ro, err := newRuntimeObjectAt(bp, p.(*Blueprint), depth)
		if err != nil {
			if errors.Is(err, ErrDepthExceeded) {
				return nil, err
			}
			return nil, nil
		}
		return ro, nil
	}

	return reflect.New(reflect.ValueOf(p).Elem().Type()).Interface().(Object), nil
}

// Add registers object prototypes. Pre-validates empty types before any insertion.
// Returns on first failure; inputs before the error are registered (bounded partial state).
//
// has() walks the parent chain so a child cannot silently shadow a prototype registered
// in a parent. Local Set collisions surface the same error.
//
// A populated *Blueprint (Type != "") is rejected: Blueprint's ObjectType is always
// "astral.blueprint", so Add would store the runtime Blueprint under the prototype slot
// and poison `New("astral.blueprint")` — use RegisterBlueprint for runtime registration.
func (bp *Blueprints) Add(object ...Object) error {
	for _, o := range object {
		if len(o.ObjectType()) == 0 {
			return fmt.Errorf("object type is empty for %s", reflect.TypeOf(o))
		}
		if b, ok := o.(*Blueprint); ok && b.Type.String() != "" {
			return fmt.Errorf("Add: use RegisterBlueprint for runtime Blueprint %s", b.Type)
		}
	}
	for _, o := range object {
		name := o.ObjectType()
		if bp.has(name) {
			return fmt.Errorf("blueprint for %s already added", name)
		}
		_, ok := bp.entries.Set(name, o)
		if !ok {
			return fmt.Errorf("blueprint for %s already added", name)
		}
	}
	return nil
}

// OrderedBlueprints returns all registered type names in dependency order. Walks
// the parent chain. At each level: non-alias compile-time prototypes first
// (alpha-sorted), then aliases (alpha-sorted, leaves with no internal refs),
// then runtime Blueprints topo-sorted by referencedType.
//
// Aliases precede runtime Blueprints so that a Blueprint's RefSpec to an alias
// resolves on the peer when replayed in this order. An entry classifies as an
// alias when it's a stored alias-kind Blueprint OR a compile-time prototype that
// satisfies PrimitiveAlias.
//
// Names that appear in both the local entries and the parent chain (parent-add-after-child
// shadow) are emitted once, with the parent occurrence preserved.
func (bp *Blueprints) OrderedBlueprints() []string {
	var out []string
	seen := map[string]bool{}
	if bp.Parent != nil {
		for _, n := range bp.Parent.OrderedBlueprints() {
			if !seen[n] {
				out = append(out, n)
				seen[n] = true
			}
		}
	}

	var proto []string
	var aliases []string
	var runtime []*Blueprint
	for name, obj := range bp.entries.Clone() {
		if seen[name] {
			continue
		}
		switch classify(name, obj) {
		case kindRuntimeBP:
			b := obj.(*Blueprint)
			if b.Kind() == BlueprintKindAlias {
				aliases = append(aliases, name)
				continue
			}
			runtime = append(runtime, b)
		case kindAliasProto:
			aliases = append(aliases, name)
		case kindStructProto:
			proto = append(proto, name)
		}
	}

	sort.Strings(proto)
	sort.Strings(aliases)
	out = append(out, proto...)
	out = append(out, aliases...)
	// note: cycle error is silenced here so the public []string signature stays unchanged;
	// AllBlueprints surfaces the same condition via its error return for callers that need it.
	ordered, _ := orderBlueprintsByReference(runtime)
	out = append(out, ordered...)
	return out
}

// orderBlueprintsByReference returns blueprint names ordered (Kahn-style topological sort)
// so each precedes any that references it. References outside the input set are treated as
// satisfied. Alpha tie-break. Returns ErrBlueprintCycle if the input contains a reference
// cycle — the names slice still contains every input (cycle nodes appended in alpha order
// after the partial sort) so callers can replay best-effort, but the error makes the cycle
// observable instead of a silent fall-through.
func orderBlueprintsByReference(bps []*Blueprint) ([]string, error) {
	if len(bps) == 0 {
		return nil, nil
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
			ref := f.Spec.ReferencedType()
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
	var cycleErr error
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
			var stuck []string
			for _, n := range names {
				if !emitted[n] {
					stuck = append(stuck, n)
					out = append(out, n)
					emitted[n] = true
				}
			}
			cycleErr = fmt.Errorf("%w: %v", ErrBlueprintCycle, stuck)
			break
		}
	}
	return out, cycleErr
}

// Register is the canonical entry point for runtime registration on the default Blueprints.
// Register stores a runtime *Blueprint (struct or alias kind) on the default Blueprints.
// Returns the content-addressed ObjectID of the stored descriptor.
func Register(o Object) (*ObjectID, error) {
	return defaultBlueprints.Register(o)
}

// Register is the canonical entry point. Returns ErrBlueprintInvalid if o is not a
// *Blueprint. The kind-specific typed method RegisterBlueprint remains for callers that
// already hold the concrete type.
func (bp *Blueprints) Register(o Object) (*ObjectID, error) {
	v, ok := o.(*Blueprint)
	if !ok {
		return nil, fmt.Errorf("%w: Register: want *Blueprint, got %T", ErrBlueprintInvalid, o)
	}
	return bp.RegisterBlueprint(v)
}

// GetBlueprint returns the runtime *Blueprint for typeName, or nil.
// Compile-time prototypes are not returned.
func GetBlueprint(typeName string) *Blueprint {
	return defaultBlueprints.GetBlueprint(typeName)
}

// RegisterBlueprint stores a runtime Blueprint after validation and returns its
// content-addressed ObjectID. Accepts both struct-kind (Fields) and alias-kind
// (Underlying) Blueprints; the kind is read off b.Kind() and validated accordingly.
// Type must not collide with any compile-time prototype or previously registered
// Blueprint — they share one map.
//
// Caller must not mutate b after this call. The registry stores the pointer as-is;
// mutations propagate to every RuntimeObject and orphan the returned ObjectID.
//
// Limitations:
//
//   - blueprint divergence between peers under the same Type is not detected on the wire;
//     a second peer's incompatible blueprint returns ErrAlreadyRegistered with no detail.
//   - parent-chain race — `has()` walks the chain but `Set()` is local-only;
//     insignificant for v1 since production targets DefaultBlueprints.
//   - duplicate Register of an identical blueprint fails instead of being idempotent.
//   - register is not identity-gated; any peer can squat a name.
func (bp *Blueprints) RegisterBlueprint(b *Blueprint) (*ObjectID, error) {
	if err := validateBlueprint(b); err != nil {
		return nil, err
	}

	typeName := b.Type.String()
	if bp.has(typeName) {
		return nil, fmt.Errorf("%w: %s", ErrAlreadyRegistered, typeName)
	}

	// why: validate that referenced types are reachable through this registry tree so
	// New() can construct a working RuntimeObject. Struct-kind walks fields; alias-kind
	// checks Underlying reachability (the primitive allowlist passed validateBlueprint,
	// but an isolated NewBlueprints(nil) without a parent chain doesn't have the
	// primitive prototype to materialize — finding registry/01).
	if b.Kind() == BlueprintKindStruct {
		if err := bp.validateReferences(b); err != nil {
			return nil, err
		}
	} else if !bp.has(b.Underlying.String()) {
		return nil, fmt.Errorf("%w: alias %s underlying %s not reachable",
			ErrBlueprintInvalid, b.Type, b.Underlying)
	}

	// why: defensive copy so a caller mutating `b` after RegisterBlueprint can't retroactively
	// change every constructed *RuntimeObject's schema or orphan the returned ObjectID. The
	// copy is shallow at the Spec level — Specs are small value carriers and the caller
	// generally constructs them as struct literals, so they aren't aliased through hidden
	// pointers.
	stored := cloneBlueprint(b)
	_, ok := bp.entries.Set(typeName, stored)
	if !ok {
		// note: raced with another caller registering the same type
		return nil, fmt.Errorf("%w: %s", ErrAlreadyRegistered, typeName)
	}

	return ResolveObjectID(stored)
}

// cloneBlueprint returns a deep-enough copy of bp to insulate the registry from caller
// mutation. Type and Underlying are value-typed (String16); Fields is reallocated; each
// Spec is shallow-copied via the per-kind copy helper.
func cloneBlueprint(bp *Blueprint) *Blueprint {
	out := &Blueprint{Type: bp.Type, Underlying: bp.Underlying}
	if len(bp.Fields) > 0 {
		out.Fields = make([]Field, len(bp.Fields))
		for i, f := range bp.Fields {
			out.Fields[i] = Field{Name: f.Name, Spec: cloneSpec(f.Spec)}
		}
	}
	return out
}

// cloneSpec returns a new pointer-to-value copy of s. Every Spec carrier in this package
// is a small struct with String16/Uint32 fields — no nested pointers — so shallow value
// copy is sufficient.
func cloneSpec(s Spec) Spec {
	switch v := s.(type) {
	case *PrimitiveSpec:
		c := *v
		return &c
	case *RefSpec:
		c := *v
		return &c
	case *SliceSpec:
		c := *v
		return &c
	case *ArraySpec:
		c := *v
		return &c
	case *MapSpec:
		c := *v
		return &c
	case *PtrSpec:
		c := *v
		return &c
	case *ObjectSpec:
		c := *v
		return &c
	}
	return s
}

// GetBlueprint returns the runtime Blueprint for typeName, or nil. Compile-time prototypes
// live under "astral.blueprint", never under their own runtime Type, so they return nil.
func (bp *Blueprints) GetBlueprint(typeName string) *Blueprint {
	o, ok := bp.entries.Get(typeName)
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
	if _, ok := bp.entries.Get(typeName); ok {
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
		ref := f.Spec.ReferencedType()
		if ref == "" || bp.has(ref) {
			continue
		}
		return fmt.Errorf("%w: field %q references unregistered type %s",
			ErrBlueprintInvalid, f.Name, ref)
	}
	return nil
}

// AllBlueprints returns every runtime Blueprint (struct kind + alias kind) ordered for sync
// replay: aliases first (alpha within each parent level), then struct-kind compile-time
// prototypes (alpha), then struct-kind runtime Blueprints topo-sorted by reference.
// Parent-chain entries precede local ones. Per-entry derivation failures (BlueprintOf/AliasOf)
// are aggregated into err; the returned slice contains the successful entries.
func (bp *Blueprints) AllBlueprints() ([]*Blueprint, error) {
	var out []*Blueprint
	var errs []error
	// why: dedupe by Type so a parent-after-child shadow (registry/02) doesn't emit two
	// entries for the same name. Parent occurrence wins; local entries with a colliding
	// name are skipped.
	seen := map[string]bool{}
	if bp.Parent != nil {
		parent, perr := bp.Parent.AllBlueprints()
		for _, b := range parent {
			if seen[b.Type.String()] {
				continue
			}
			out = append(out, b)
			seen[b.Type.String()] = true
		}
		if perr != nil {
			errs = append(errs, perr)
		}
	}

	// why: single Clone() walk bucketed via classify; alias-kind Blueprints are pulled out
	// of the runtime bucket so sync can replay aliases before any struct-kind Blueprint
	// RefSpec-ing them.
	var aliases []*Blueprint
	var proto []*Blueprint
	var runtime []*Blueprint
	for name, obj := range bp.entries.Clone() {
		if seen[name] {
			continue
		}
		switch classify(name, obj) {
		case kindRuntimeBP:
			b := obj.(*Blueprint)
			if b.Kind() == BlueprintKindAlias {
				aliases = append(aliases, b)
			} else {
				runtime = append(runtime, b)
			}
		case kindAliasProto:
			derived, err := BlueprintOf(obj)
			if err != nil {
				errs = append(errs, fmt.Errorf("alias %s: %w", name, err))
				continue
			}
			aliases = append(aliases, derived)
		case kindStructProto:
			derived, err := BlueprintOf(obj)
			if err != nil {
				errs = append(errs, fmt.Errorf("blueprint %s: %w", name, err))
				continue
			}
			proto = append(proto, derived)
		}
	}

	sort.Slice(aliases, func(i, j int) bool {
		return aliases[i].Type.String() < aliases[j].Type.String()
	})
	out = append(out, aliases...)

	sort.Slice(proto, func(i, j int) bool {
		return proto[i].Type.String() < proto[j].Type.String()
	})
	out = append(out, proto...)

	ordered, cycleErr := orderBlueprintsByReference(runtime)
	if cycleErr != nil {
		errs = append(errs, cycleErr)
	}
	for _, name := range ordered {
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
