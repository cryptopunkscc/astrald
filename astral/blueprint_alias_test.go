package astral

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"strings"
	"testing"
)

// aliasesOnly filters a Blueprint slice down to alias-kind entries. Replaces the old
// AllBlueprintAliases method for tests that asserted on derived/registered aliases only.
func aliasesOnly(in []*Blueprint) []*Blueprint {
	var out []*Blueprint
	for _, b := range in {
		if b.Kind() == BlueprintKindAlias {
			out = append(out, b)
		}
	}
	return out
}

// getAlias returns the registered alias-kind Blueprint under typeName, or nil — a thin
// wrapper that replaces the old GetAlias method.
func getAlias(bps *Blueprints, typeName string) *Blueprint {
	b := bps.GetBlueprint(typeName)
	if b == nil || b.Kind() != BlueprintKindAlias {
		return nil
	}
	return b
}

func TestRegisterAlias_Success(t *testing.T) {
	bps := NewBlueprints(DefaultBlueprints())
	a := NewBlueprintAlias("test.alias_ok", "uint8")

	id, err := bps.RegisterBlueprint(a)
	if err != nil {
		t.Fatal(err)
	}
	if id == nil {
		t.Fatal("expected non-nil ObjectID")
	}
	if got := getAlias(bps, "test.alias_ok"); got == nil {
		t.Fatal("alias-kind lookup returned nil for registered Type")
	}
}

func TestRegisterAlias_EmptyType(t *testing.T) {
	bps := NewBlueprints(DefaultBlueprints())
	a := NewBlueprintAlias("", "uint8")
	_, err := bps.RegisterBlueprint(a)
	if !errors.Is(err, ErrBlueprintInvalid) {
		t.Fatalf("want ErrBlueprintInvalid, got %v", err)
	}
}

func TestRegisterAlias_BadUnderlying(t *testing.T) {
	bps := NewBlueprints(DefaultBlueprints())
	a := NewBlueprintAlias("test.alias_bad_under", "int32")
	_, err := bps.RegisterBlueprint(a)
	if !errors.Is(err, ErrBlueprintInvalid) {
		t.Fatalf("want ErrBlueprintInvalid, got %v", err)
	}
}

func TestRegisterAlias_NonASCIIType(t *testing.T) {
	bps := NewBlueprints(DefaultBlueprints())
	a := NewBlueprintAlias("test.alias_café", "uint8")
	_, err := bps.RegisterBlueprint(a)
	if !errors.Is(err, ErrBlueprintInvalid) {
		t.Fatalf("want ErrBlueprintInvalid, got %v", err)
	}
}

func TestRegisterAlias_OversizedType(t *testing.T) {
	bps := NewBlueprints(DefaultBlueprints())
	a := NewBlueprintAlias(strings.Repeat("a", MaxBlueprintNameLen+1), "uint8")
	_, err := bps.RegisterBlueprint(a)
	if !errors.Is(err, ErrBlueprintInvalid) {
		t.Fatalf("want ErrBlueprintInvalid, got %v", err)
	}
}

func TestRegisterAlias_DuplicateAlias(t *testing.T) {
	bps := NewBlueprints(DefaultBlueprints())
	a1 := NewBlueprintAlias("test.alias_dup", "uint8")
	a2 := NewBlueprintAlias("test.alias_dup", "uint16")
	if _, err := bps.RegisterBlueprint(a1); err != nil {
		t.Fatal(err)
	}
	if _, err := bps.RegisterBlueprint(a2); !errors.Is(err, ErrAlreadyRegistered) {
		t.Fatalf("want ErrAlreadyRegistered, got %v", err)
	}
}

// TestRegisterAlias_RejectsExistingPrototype pins the single-map contract: a name held
// by a compile-time prototype cannot also be registered as an alias — the prototype
// already covers it (and, if it implements PrimitiveAlias, supplies the alias via derivation).
func TestRegisterAlias_RejectsExistingPrototype(t *testing.T) {
	bps := NewBlueprints(nil)
	b := Bool(false)
	if err := bps.Add(&b); err != nil {
		t.Fatal(err)
	}
	_, err := bps.RegisterBlueprint(NewBlueprintAlias("bool", "uint8"))
	if !errors.Is(err, ErrAlreadyRegistered) {
		t.Fatalf("alias under an existing prototype name must fail, got %v", err)
	}
}

// TestAllBlueprints_DerivesAliasKindFromPrimitiveAliasPrototype pins the PrimitiveAlias-derivation
// path: a compile-time prototype that implements UnderlyingPrimitive surfaces as an
// alias-kind Blueprint via AllBlueprints without an explicit RegisterBlueprint call.
func TestAllBlueprints_DerivesAliasKindFromPrimitiveAliasPrototype(t *testing.T) {
	bps := NewBlueprints(nil)
	if err := bps.Add(newPrimitiveAliasMode()); err != nil {
		t.Fatal(err)
	}
	all, err := bps.AllBlueprints()
	if err != nil {
		t.Fatal(err)
	}
	got := aliasesOnly(all)
	if len(got) != 1 {
		t.Fatalf("want 1 derived alias, got %d", len(got))
	}
	if got[0].Type.String() != "test.aliasable_mode" || got[0].Underlying.String() != "uint8" {
		t.Fatalf("derived alias has wrong shape: %+v", got[0])
	}
}

// TestNew_PrimitiveAliasPrototypeReturnsGoValue pins that local New() returns the typed Go
// value for an PrimitiveAlias prototype (not *RuntimeObject) — sync handles the alias view; the
// originating peer always works with the Go type.
func TestNew_PrimitiveAliasPrototypeReturnsGoValue(t *testing.T) {
	bps := NewBlueprints(nil)
	if err := bps.Add(newPrimitiveAliasMode()); err != nil {
		t.Fatal(err)
	}
	obj := bps.New("test.aliasable_mode")
	if _, ok := obj.(*aliasableMode); !ok {
		t.Fatalf("expected *aliasableMode from prototype, got %T", obj)
	}
}

// aliasableMode is a test stand-in for app-level types like nearby.Mode: a Uint8 newtype
// that satisfies PrimitiveAlias so the registry can derive its alias-kind Blueprint.
type aliasableMode Uint8

func newPrimitiveAliasMode() *aliasableMode { m := aliasableMode(0); return &m }

// astral:blueprint-ignore
func (*aliasableMode) ObjectType() string                   { return "test.aliasable_mode" }
func (*aliasableMode) UnderlyingPrimitive() string          { return "uint8" }
func (m *aliasableMode) WriteTo(w io.Writer) (int64, error) { return (*Uint8)(m).WriteTo(w) }
func (m *aliasableMode) ReadFrom(r io.Reader) (int64, error) {
	return (*Uint8)(m).ReadFrom(r)
}

func TestNew_AliasKindReturnsRuntimeObject(t *testing.T) {
	bps := NewBlueprints(DefaultBlueprints())
	if _, err := bps.RegisterBlueprint(NewBlueprintAlias("test.alias_new", "uint32")); err != nil {
		t.Fatal(err)
	}
	obj := bps.New("test.alias_new")
	if obj == nil {
		t.Fatal("New returned nil for registered alias")
	}
	ra, ok := obj.(*RuntimeObject)
	if !ok {
		t.Fatalf("expected *RuntimeObject, got %T", obj)
	}
	if ra.ObjectType() != "test.alias_new" {
		t.Fatalf("want type test.alias_new, got %s", ra.ObjectType())
	}
	if _, ok := ra.Underlying().(*Uint32); !ok {
		t.Fatalf("expected underlying *Uint32, got %T", ra.Underlying())
	}
}

func TestAliasKind_RoundTrip(t *testing.T) {
	bps := NewBlueprints(DefaultBlueprints())
	if _, err := bps.RegisterBlueprint(NewBlueprintAlias("test.alias_rt", "uint8")); err != nil {
		t.Fatal(err)
	}

	src := bps.New("test.alias_rt").(*RuntimeObject)
	v := Uint8(7)
	src.value = &v

	var buf bytes.Buffer
	n, err := src.WriteTo(&buf)
	if err != nil {
		t.Fatal(err)
	}
	if n != 1 {
		t.Fatalf("uint8 alias must write 1 byte, got %d", n)
	}

	dst := bps.New("test.alias_rt").(*RuntimeObject)
	_, err = dst.ReadFrom(&buf)
	if err != nil {
		t.Fatal(err)
	}
	got, ok := dst.Underlying().(*Uint8)
	if !ok || uint8(*got) != 7 {
		t.Fatalf("round-trip lost value: got %#v", dst.Underlying())
	}
}

// TestBlueprintRef_ToAlias_Resolves pins the registration-time closure invariant: a
// struct-kind Blueprint that RefSpecs an alias-kind Blueprint validates closure against
// the shared registry map.
func TestBlueprintRef_ToAlias_Resolves(t *testing.T) {
	bps := NewBlueprints(DefaultBlueprints())
	if _, err := bps.RegisterBlueprint(NewBlueprintAlias("test.alias_target", "uint8")); err != nil {
		t.Fatal(err)
	}
	bp := NewBlueprint("test.alias_referrer",
		Field{Name: "m", Spec: &RefSpec{Type: "test.alias_target"}},
	)
	if _, err := bps.RegisterBlueprint(bp); err != nil {
		t.Fatalf("RefSpec to alias should resolve, got %v", err)
	}
}

func TestOrderedTypes_AliasesBeforeRuntimeBlueprints(t *testing.T) {
	// why: alias-kind registration now requires the underlying primitive to be reachable
	// (registry/01); parent through DefaultBlueprints provides the uint8 prototype.
	bps := NewBlueprints(DefaultBlueprints())
	if _, err := bps.RegisterBlueprint(NewBlueprintAlias("test.alias_z", "uint8")); err != nil {
		t.Fatal(err)
	}
	if _, err := bps.RegisterBlueprint(NewBlueprint("test.bp_a",
		Field{Name: "r", Spec: &RefSpec{Type: "test.alias_z"}},
	)); err != nil {
		t.Fatal(err)
	}
	got := bps.OrderedBlueprints()
	mustPrecede(t, got, "test.alias_z", "test.bp_a")
}

// TestOrderedTypes_PrimitiveAliasPrototypeBucketsAsAlias pins that a compile-time prototype
// implementing PrimitiveAlias is emitted in the alias bucket (between proto and runtime
// blueprints) rather than the proto bucket — so a runtime Blueprint that RefSpecs it
// resolves on replay.
func TestOrderedTypes_PrimitiveAliasPrototypeBucketsAsAlias(t *testing.T) {
	bps := NewBlueprints(nil)
	if err := bps.Add(newPrimitiveAliasMode()); err != nil {
		t.Fatal(err)
	}
	if _, err := bps.RegisterBlueprint(NewBlueprint("test.refs_aliasable",
		Field{Name: "m", Spec: &RefSpec{Type: "test.aliasable_mode"}},
	)); err != nil {
		t.Fatal(err)
	}
	got := bps.OrderedBlueprints()
	mustPrecede(t, got, "test.aliasable_mode", "test.refs_aliasable")
}

func TestAllBlueprints_AliasesAlphaWithinLevel(t *testing.T) {
	bps := NewBlueprints(DefaultBlueprints())
	for _, n := range []string{"test.alias_c", "test.alias_a", "test.alias_b"} {
		if _, err := bps.RegisterBlueprint(NewBlueprintAlias(n, "uint8")); err != nil {
			t.Fatal(err)
		}
	}
	// note: AllBlueprints aggregates per-entry derivation failures (DefaultBlueprints
	// contains primitive prototypes that can't be reflected into struct-kind Blueprints).
	// We only assert on the test-scoped aliases below; the aggregate error is expected.
	all, _ := bps.AllBlueprints()
	var got []*Blueprint
	for _, a := range aliasesOnly(all) {
		if strings.HasPrefix(a.Type.String(), "test.alias_") {
			got = append(got, a)
		}
	}
	if len(got) != 3 {
		t.Fatalf("want 3, got %d", len(got))
	}
	want := []string{"test.alias_a", "test.alias_b", "test.alias_c"}
	for i, a := range got {
		if a.Type.String() != want[i] {
			t.Fatalf("idx %d: want %s, got %s", i, want[i], a.Type)
		}
	}
}

func TestAllBlueprints_AliasesParentChain(t *testing.T) {
	parent := NewBlueprints(DefaultBlueprints())
	if _, err := parent.RegisterBlueprint(NewBlueprintAlias("test.alias_parent", "uint8")); err != nil {
		t.Fatal(err)
	}
	child := NewBlueprints(parent)
	if _, err := child.RegisterBlueprint(NewBlueprintAlias("test.alias_child", "uint16")); err != nil {
		t.Fatal(err)
	}
	// note: same aggregate-error rationale as TestAllBlueprints_AliasesAlphaWithinLevel.
	all, _ := child.AllBlueprints()
	// Filter to the test-scoped aliases — DefaultBlueprints contributes its own aliases
	// via the parent chain now that alias registration requires reachable primitives.
	var got []*Blueprint
	for _, a := range aliasesOnly(all) {
		if strings.HasPrefix(a.Type.String(), "test.alias_") {
			got = append(got, a)
		}
	}
	if len(got) != 2 {
		t.Fatalf("want 2, got %d", len(got))
	}
	if got[0].Type.String() != "test.alias_parent" {
		t.Fatalf("parent must precede child: %v", got)
	}
}

// TestAllBlueprints_BadPrimitiveAliasPrimitiveAggregated pins that an PrimitiveAlias prototype reporting
// an unallowed underlying primitive is skipped with an aggregated error; the rest of the
// derivation still succeeds.
func TestAllBlueprints_BadPrimitiveAliasPrimitiveAggregated(t *testing.T) {
	bps := NewBlueprints(nil)
	if err := bps.Add(newPrimitiveAliasMode(), &badPrimitiveAlias{}); err != nil {
		t.Fatal(err)
	}
	all, err := bps.AllBlueprints()
	if err == nil {
		t.Fatal("expected aggregated error for bad underlying")
	}
	got := aliasesOnly(all)
	if len(got) != 1 || got[0].Type.String() != "test.aliasable_mode" {
		t.Fatalf("expected only the good alias entry, got %d entries", len(got))
	}
}

// badPrimitiveAlias returns an underlying primitive that isn't on the allowlist.
type badPrimitiveAlias struct{}

// astral:blueprint-ignore
func (*badPrimitiveAlias) ObjectType() string                  { return "test.bad_aliasable" }
func (*badPrimitiveAlias) UnderlyingPrimitive() string         { return "int32" }
func (*badPrimitiveAlias) WriteTo(w io.Writer) (int64, error)  { return 0, nil }
func (*badPrimitiveAlias) ReadFrom(r io.Reader) (int64, error) { return 0, nil }

// TestEncodeDecode_AliasCrossesBlueprints pins the wire path: a value of an PrimitiveAlias
// type, encoded via the polymorphic envelope on the originating peer (which has the Go
// type), decodes on a peer that has ONLY the alias-kind Blueprint as *RuntimeObject with
// the underlying byte intact.
func TestEncodeDecode_AliasCrossesBlueprints(t *testing.T) {
	// originating peer: typed Go value
	local := NewBlueprints(DefaultBlueprints())
	if err := local.Add(newPrimitiveAliasMode()); err != nil {
		t.Fatal(err)
	}
	src := aliasableMode(7)
	buf, err := EncodeBytes(&src, WithBlueprints(local))
	if err != nil {
		t.Fatal(err)
	}

	// remote peer: only the schema descriptor (alias-kind Blueprint), no Go type
	remote := NewBlueprints(DefaultBlueprints())
	if _, err := remote.RegisterBlueprint(NewBlueprintAlias("test.aliasable_mode", "uint8")); err != nil {
		t.Fatal(err)
	}
	got, _, err := Decode(bytes.NewReader(buf), WithBlueprints(remote))
	if err != nil {
		t.Fatal(err)
	}
	ra, ok := got.(*RuntimeObject)
	if !ok {
		t.Fatalf("remote decode: want *RuntimeObject, got %T", got)
	}
	if ra.ObjectType() != "test.aliasable_mode" {
		t.Fatalf("ObjectType drift: %s", ra.ObjectType())
	}
	u, ok := ra.Underlying().(*Uint8)
	if !ok || uint8(*u) != 7 {
		t.Fatalf("payload drift: %#v", ra.Underlying())
	}
}

// TestAliasKind_JSONRoundTrip pins that the JSON shape is the bare underlying value
// (no envelope, no wrapper) and that UnmarshalJSON populates a bound carrier.
func TestAliasKind_JSONRoundTrip(t *testing.T) {
	bps := NewBlueprints(DefaultBlueprints())
	if _, err := bps.RegisterBlueprint(NewBlueprintAlias("test.alias_json", "uint16")); err != nil {
		t.Fatal(err)
	}
	src := bps.New("test.alias_json").(*RuntimeObject)
	v := Uint16(42)
	src.value = &v

	data, err := json.Marshal(src)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "42" {
		t.Fatalf("alias JSON should be bare value, got %s", data)
	}

	dst := bps.New("test.alias_json").(*RuntimeObject)
	if err := json.Unmarshal(data, dst); err != nil {
		t.Fatal(err)
	}
	got, ok := dst.Underlying().(*Uint16)
	if !ok || uint16(*got) != 42 {
		t.Fatalf("JSON round-trip lost value: %#v", dst.Underlying())
	}
}

// TestBlueprintOf_RejectsNonStructNonPrimitiveAlias pins that BlueprintOf on a non-struct,
// non-PrimitiveAlias type (here Bool, which is `type Bool bool`) fails — there's no shape to
// derive.
func TestBlueprintOf_RejectsNonStructNonPrimitiveAlias(t *testing.T) {
	b := Bool(false)
	if _, err := BlueprintOf(&b); err == nil {
		t.Fatal("expected error for non-struct non-PrimitiveAlias, got nil")
	}
}

// TestBlueprintOf_RejectsBadPrimitiveAliasUnderlying pins that an PrimitiveAlias returning an
// unallowed primitive is caught by BlueprintOf via validateBlueprint.
func TestBlueprintOf_RejectsBadPrimitiveAliasUnderlying(t *testing.T) {
	_, err := BlueprintOf(&badPrimitiveAlias{})
	if !errors.Is(err, ErrBlueprintInvalid) {
		t.Fatalf("want ErrBlueprintInvalid for bad underlying, got %v", err)
	}
}

// TestBlueprintOf_PrimitiveAliasProducesAliasKind pins the happy path: an PrimitiveAlias prototype
// is derived as an alias-kind *Blueprint.
func TestBlueprintOf_PrimitiveAliasProducesAliasKind(t *testing.T) {
	got, err := BlueprintOf(newPrimitiveAliasMode())
	if err != nil {
		t.Fatal(err)
	}
	if got.Kind() != BlueprintKindAlias {
		t.Fatalf("want alias kind, got %v", got.Kind())
	}
	if got.Type.String() != "test.aliasable_mode" || got.Underlying.String() != "uint8" {
		t.Fatalf("derived alias wrong shape: %+v", got)
	}
}

func TestRegister_DispatchesByType(t *testing.T) {
	bps := NewBlueprints(DefaultBlueprints())

	bpID, err := bps.Register(NewBlueprint("test.register_bp",
		Field{Name: "n", Spec: &PrimitiveSpec{PrimitiveType: "uint32"}},
	))
	if err != nil || bpID == nil {
		t.Fatalf("Register(struct-kind Blueprint) failed: %v", err)
	}
	if bps.GetBlueprint("test.register_bp") == nil {
		t.Fatal("Register(struct-kind) did not store the Blueprint")
	}

	aID, err := bps.Register(NewBlueprintAlias("test.register_alias", "uint16"))
	if err != nil || aID == nil {
		t.Fatalf("Register(alias-kind Blueprint) failed: %v", err)
	}
	if getAlias(bps, "test.register_alias") == nil {
		t.Fatal("Register(alias-kind) did not store the alias Blueprint")
	}

	// Anything else must be rejected with ErrBlueprintInvalid.
	b := Bool(false)
	_, err = bps.Register(&b)
	if !errors.Is(err, ErrBlueprintInvalid) {
		t.Fatalf("Register(other) should reject with ErrBlueprintInvalid, got %v", err)
	}
}

// TestRefSpec_ToAlias_EndToEnd pins the WithBlueprints plumbing: a Blueprint with a
// RefSpec to an PrimitiveAlias type encodes through the originating peer's Go value and the
// remote peer decodes via the alias entry in its registry. Before objectReader.bps was
// threaded through Decode, this failed at "blueprint not found" because readField used
// the package-level (default) Blueprints regardless of WithBlueprints.
func TestRefSpec_ToAlias_EndToEnd(t *testing.T) {
	// local: typed Go value + a Blueprint referencing it
	local := NewBlueprints(DefaultBlueprints())
	if err := local.Add(newPrimitiveAliasMode()); err != nil {
		t.Fatal(err)
	}
	if _, err := local.RegisterBlueprint(NewBlueprint("test.envelope",
		Field{Name: "m", Spec: &RefSpec{Type: "test.aliasable_mode"}},
	)); err != nil {
		t.Fatal(err)
	}
	src := local.New("test.envelope").(*RuntimeObject)
	mode := aliasableMode(3)
	if err := src.Set("m", &mode); err != nil {
		t.Fatal(err)
	}
	buf, err := EncodeBytes(src, WithBlueprints(local))
	if err != nil {
		t.Fatal(err)
	}

	// remote: alias + same Blueprint, no Go type for Mode
	remote := NewBlueprints(DefaultBlueprints())
	if _, err := remote.RegisterBlueprint(NewBlueprintAlias("test.aliasable_mode", "uint8")); err != nil {
		t.Fatal(err)
	}
	if _, err := remote.RegisterBlueprint(NewBlueprint("test.envelope",
		Field{Name: "m", Spec: &RefSpec{Type: "test.aliasable_mode"}},
	)); err != nil {
		t.Fatal(err)
	}
	got, _, err := Decode(bytes.NewReader(buf), WithBlueprints(remote))
	if err != nil {
		t.Fatal(err)
	}
	ro, ok := got.(*RuntimeObject)
	if !ok {
		t.Fatalf("want *RuntimeObject, got %T", got)
	}
	field := ro.Get("m")
	ra, ok := field.(*RuntimeObject)
	if !ok {
		t.Fatalf("field m: want *RuntimeObject, got %T", field)
	}
	u, ok := ra.Underlying().(*Uint8)
	if !ok || uint8(*u) != 3 {
		t.Fatalf("field m underlying drift: %#v", ra.Underlying())
	}
}

// TestPtrSpec_ToAlias_EndToEnd mirrors TestRefSpec_ToAlias_EndToEnd for the PtrSpec
// branch of readField: a Blueprint with a PtrSpec to an PrimitiveAlias type must also resolve
// via the caller's WithBlueprints registry, not the package-level default.
func TestPtrSpec_ToAlias_EndToEnd(t *testing.T) {
	// local: typed Go value + a Blueprint referencing it via PtrSpec
	local := NewBlueprints(DefaultBlueprints())
	if err := local.Add(newPrimitiveAliasMode()); err != nil {
		t.Fatal(err)
	}
	if _, err := local.RegisterBlueprint(NewBlueprint("test.ptr_envelope",
		Field{Name: "m", Spec: &PtrSpec{Type: "test.aliasable_mode"}},
	)); err != nil {
		t.Fatal(err)
	}
	src := local.New("test.ptr_envelope").(*RuntimeObject)
	mode := aliasableMode(3)
	if err := src.Set("m", &mode); err != nil {
		t.Fatal(err)
	}
	buf, err := EncodeBytes(src, WithBlueprints(local))
	if err != nil {
		t.Fatal(err)
	}

	// remote: alias + same Blueprint, no Go type for Mode
	remote := NewBlueprints(DefaultBlueprints())
	if _, err := remote.RegisterBlueprint(NewBlueprintAlias("test.aliasable_mode", "uint8")); err != nil {
		t.Fatal(err)
	}
	if _, err := remote.RegisterBlueprint(NewBlueprint("test.ptr_envelope",
		Field{Name: "m", Spec: &PtrSpec{Type: "test.aliasable_mode"}},
	)); err != nil {
		t.Fatal(err)
	}
	got, _, err := Decode(bytes.NewReader(buf), WithBlueprints(remote))
	if err != nil {
		t.Fatal(err)
	}
	ro, ok := got.(*RuntimeObject)
	if !ok {
		t.Fatalf("want *RuntimeObject, got %T", got)
	}
	field := ro.Get("m")
	ra, ok := field.(*RuntimeObject)
	if !ok {
		t.Fatalf("field m: want *RuntimeObject, got %T", field)
	}
	u, ok := ra.Underlying().(*Uint8)
	if !ok || uint8(*u) != 3 {
		t.Fatalf("field m underlying drift: %#v", ra.Underlying())
	}
}

// TestRuntimeSlice_RoundTrip_HonorsCustomBlueprints pins end-to-end Path A for the SliceSpec
// branch: a *RuntimeSlice whose element type is a Blueprint registered ONLY in a custom
// registry must encode and decode through EncodeBytes/Decode when WithBlueprints carries the
// custom registry. Before the read fix, isRuntimeBlueprintType / readRuntimeBlueprintPtr
// consulted defaultBlueprints — the slow path was never engaged for custom-only elements,
// and the fast path silently no-oped per element. Before the write fix, writeField
// re-validated SliceSpec types against defaultBlueprints, so encode also failed.
func TestRuntimeSlice_RoundTrip_HonorsCustomBlueprints(t *testing.T) {
	custom := NewBlueprints(DefaultBlueprints())
	if _, err := custom.RegisterBlueprint(NewBlueprint("test.slice_item_custom",
		Field{Name: "v", Spec: &PrimitiveSpec{PrimitiveType: "uint32"}},
	)); err != nil {
		t.Fatal(err)
	}
	if _, err := custom.RegisterBlueprint(NewBlueprint("test.slice_envelope_custom",
		Field{Name: "items", Spec: &SliceSpec{Type: "test.slice_item_custom"}},
	)); err != nil {
		t.Fatal(err)
	}

	src := custom.New("test.slice_envelope_custom").(*RuntimeObject)
	items, err := NewRuntimeSliceWith(custom, "test.slice_item_custom")
	if err != nil {
		t.Fatal(err)
	}
	for _, v := range []uint32{42, 99} {
		elem := custom.New("test.slice_item_custom").(*RuntimeObject)
		if err := elem.Set("v", v); err != nil {
			t.Fatal(err)
		}
		if err := items.Append(elem); err != nil {
			t.Fatal(err)
		}
	}
	if err := src.Set("items", items); err != nil {
		t.Fatal(err)
	}

	buf, err := EncodeBytes(src, WithBlueprints(custom))
	if err != nil {
		t.Fatalf("encode: %v", err)
	}

	got, _, err := Decode(bytes.NewReader(buf), WithBlueprints(custom))
	if err != nil {
		t.Fatalf("decode: %v", err)
	}
	dst, ok := got.(*RuntimeObject)
	if !ok {
		t.Fatalf("want *RuntimeObject, got %T", got)
	}
	rs, ok := dst.Get("items").(*RuntimeSlice)
	if !ok || rs.Len() != 2 {
		t.Fatalf("decoded items: want 2 elements, got %#v", dst.Get("items"))
	}
	for i, want := range []uint32{42, 99} {
		elem, ok := rs.At(i).(*RuntimeObject)
		if !ok {
			t.Fatalf("element %d: want *RuntimeObject, got %T", i, rs.At(i))
		}
		got, ok := elem.Get("v").(*Uint32)
		if !ok || uint32(*got) != want {
			t.Fatalf("element %d: v=%v, want %d", i, elem.Get("v"), want)
		}
	}
}
