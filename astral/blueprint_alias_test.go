package astral

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"strings"
	"testing"
)

func TestRegisterAlias_Success(t *testing.T) {
	bps := NewBlueprints(DefaultBlueprints())
	a := NewBlueprintAlias("test.alias_ok", "uint8")

	id, err := bps.RegisterAlias(a)
	if err != nil {
		t.Fatal(err)
	}
	if id == nil {
		t.Fatal("expected non-nil ObjectID")
	}
	if got := bps.GetAlias("test.alias_ok"); got == nil {
		t.Fatal("GetAlias returned nil for registered Type")
	}
}

func TestRegisterAlias_EmptyType(t *testing.T) {
	bps := NewBlueprints(DefaultBlueprints())
	a := NewBlueprintAlias("", "uint8")
	_, err := bps.RegisterAlias(a)
	if !errors.Is(err, ErrBlueprintInvalid) {
		t.Fatalf("want ErrBlueprintInvalid, got %v", err)
	}
}

func TestRegisterAlias_BadUnderlying(t *testing.T) {
	bps := NewBlueprints(DefaultBlueprints())
	a := NewBlueprintAlias("test.alias_bad_under", "int32")
	_, err := bps.RegisterAlias(a)
	if !errors.Is(err, ErrBlueprintInvalid) {
		t.Fatalf("want ErrBlueprintInvalid, got %v", err)
	}
}

func TestRegisterAlias_NonASCIIType(t *testing.T) {
	bps := NewBlueprints(DefaultBlueprints())
	a := NewBlueprintAlias("test.alias_café", "uint8")
	_, err := bps.RegisterAlias(a)
	if !errors.Is(err, ErrBlueprintInvalid) {
		t.Fatalf("want ErrBlueprintInvalid, got %v", err)
	}
}

func TestRegisterAlias_OversizedType(t *testing.T) {
	bps := NewBlueprints(DefaultBlueprints())
	a := NewBlueprintAlias(strings.Repeat("a", MaxBlueprintNameLen+1), "uint8")
	_, err := bps.RegisterAlias(a)
	if !errors.Is(err, ErrBlueprintInvalid) {
		t.Fatalf("want ErrBlueprintInvalid, got %v", err)
	}
}

func TestRegisterAlias_DuplicateAlias(t *testing.T) {
	bps := NewBlueprints(DefaultBlueprints())
	a1 := NewBlueprintAlias("test.alias_dup", "uint8")
	a2 := NewBlueprintAlias("test.alias_dup", "uint16")
	if _, err := bps.RegisterAlias(a1); err != nil {
		t.Fatal(err)
	}
	if _, err := bps.RegisterAlias(a2); !errors.Is(err, ErrAlreadyRegistered) {
		t.Fatalf("want ErrAlreadyRegistered, got %v", err)
	}
}

// TestRegisterAlias_RejectsExistingPrototype pins the single-map contract: a name held
// by a compile-time prototype cannot also be registered as an alias — the prototype
// already covers it (and, if it implements Aliasable, supplies the alias via derivation).
func TestRegisterAlias_RejectsExistingPrototype(t *testing.T) {
	bps := NewBlueprints(nil)
	b := Bool(false)
	if err := bps.Add(&b); err != nil {
		t.Fatal(err)
	}
	_, err := bps.RegisterAlias(NewBlueprintAlias("bool", "uint8"))
	if !errors.Is(err, ErrAlreadyRegistered) {
		t.Fatalf("alias under an existing prototype name must fail, got %v", err)
	}
}

// TestAllAliases_DerivesFromAliasablePrototype pins the Aliasable-derivation path: a
// compile-time prototype that implements UnderlyingPrimitive surfaces as a BlueprintAlias
// without an explicit RegisterAlias call.
func TestAllAliases_DerivesFromAliasablePrototype(t *testing.T) {
	bps := NewBlueprints(nil)
	if err := bps.Add(newAliasableMode()); err != nil {
		t.Fatal(err)
	}
	got, err := bps.AllAliases()
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 {
		t.Fatalf("want 1 derived alias, got %d", len(got))
	}
	if got[0].Type.String() != "test.aliasable_mode" || got[0].Underlying.String() != "uint8" {
		t.Fatalf("derived alias has wrong shape: %+v", got[0])
	}
}

// TestNew_AliasablePrototypeReturnsGoValue pins that local New() returns the typed Go
// value for an Aliasable prototype (not *RuntimeAlias) — sync handles the alias view; the
// originating peer always works with the Go type.
func TestNew_AliasablePrototypeReturnsGoValue(t *testing.T) {
	bps := NewBlueprints(nil)
	if err := bps.Add(newAliasableMode()); err != nil {
		t.Fatal(err)
	}
	obj := bps.New("test.aliasable_mode")
	if _, ok := obj.(*aliasableMode); !ok {
		t.Fatalf("expected *aliasableMode from prototype, got %T", obj)
	}
}

// aliasableMode is a test stand-in for app-level types like nearby.Mode: a Uint8 newtype
// that satisfies Aliasable so the registry can derive its BlueprintAlias.
type aliasableMode Uint8

func newAliasableMode() *aliasableMode { m := aliasableMode(0); return &m }

// astral:blueprint-ignore
func (*aliasableMode) ObjectType() string                   { return "test.aliasable_mode" }
func (*aliasableMode) UnderlyingPrimitive() string          { return "uint8" }
func (m *aliasableMode) WriteTo(w io.Writer) (int64, error) { return (*Uint8)(m).WriteTo(w) }
func (m *aliasableMode) ReadFrom(r io.Reader) (int64, error) {
	return (*Uint8)(m).ReadFrom(r)
}

func TestNew_ReturnsRuntimeAlias(t *testing.T) {
	bps := NewBlueprints(DefaultBlueprints())
	if _, err := bps.RegisterAlias(NewBlueprintAlias("test.alias_new", "uint32")); err != nil {
		t.Fatal(err)
	}
	obj := bps.New("test.alias_new")
	if obj == nil {
		t.Fatal("New returned nil for registered alias")
	}
	ra, ok := obj.(*RuntimeAlias)
	if !ok {
		t.Fatalf("expected *RuntimeAlias, got %T", obj)
	}
	if ra.ObjectType() != "test.alias_new" {
		t.Fatalf("want type test.alias_new, got %s", ra.ObjectType())
	}
	if _, ok := ra.Underlying().(*Uint32); !ok {
		t.Fatalf("expected underlying *Uint32, got %T", ra.Underlying())
	}
}

func TestRuntimeAlias_RoundTrip(t *testing.T) {
	bps := NewBlueprints(DefaultBlueprints())
	if _, err := bps.RegisterAlias(NewBlueprintAlias("test.alias_rt", "uint8")); err != nil {
		t.Fatal(err)
	}

	src := bps.New("test.alias_rt").(*RuntimeAlias)
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

	dst := bps.New("test.alias_rt").(*RuntimeAlias)
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
// Blueprint that RefSpecs an alias name must validate against the aliases map (not just
// the main Blueprints map).
func TestBlueprintRef_ToAlias_Resolves(t *testing.T) {
	bps := NewBlueprints(DefaultBlueprints())
	if _, err := bps.RegisterAlias(NewBlueprintAlias("test.alias_target", "uint8")); err != nil {
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
	bps := NewBlueprints(nil)
	if _, err := bps.RegisterAlias(NewBlueprintAlias("test.alias_z", "uint8")); err != nil {
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

// TestOrderedTypes_AliasablePrototypeBucketsAsAlias pins that a compile-time prototype
// implementing Aliasable is emitted in the alias bucket (between proto and runtime
// blueprints) rather than the proto bucket — so a runtime Blueprint that RefSpecs it
// resolves on replay.
func TestOrderedTypes_AliasablePrototypeBucketsAsAlias(t *testing.T) {
	bps := NewBlueprints(nil)
	if err := bps.Add(newAliasableMode()); err != nil {
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

func TestAllAliases_AlphaWithinLevel(t *testing.T) {
	bps := NewBlueprints(nil)
	for _, n := range []string{"test.alias_c", "test.alias_a", "test.alias_b"} {
		if _, err := bps.RegisterAlias(NewBlueprintAlias(n, "uint8")); err != nil {
			t.Fatal(err)
		}
	}
	got, err := bps.AllAliases()
	if err != nil {
		t.Fatal(err)
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

func TestAllAliases_ParentChain(t *testing.T) {
	parent := NewBlueprints(nil)
	if _, err := parent.RegisterAlias(NewBlueprintAlias("test.alias_parent", "uint8")); err != nil {
		t.Fatal(err)
	}
	child := NewBlueprints(parent)
	if _, err := child.RegisterAlias(NewBlueprintAlias("test.alias_child", "uint16")); err != nil {
		t.Fatal(err)
	}
	got, err := child.AllAliases()
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 2 {
		t.Fatalf("want 2, got %d", len(got))
	}
	if got[0].Type.String() != "test.alias_parent" {
		t.Fatalf("parent must precede child: %v", got)
	}
}

// TestAllAliases_BadAliasablePrimitiveAggregated pins that a prototype reporting an
// unallowed underlying primitive is skipped with an aggregated error — the rest of the
// derivation still succeeds.
func TestAllAliases_BadAliasablePrimitiveAggregated(t *testing.T) {
	bps := NewBlueprints(nil)
	if err := bps.Add(newAliasableMode(), &badAliasable{}); err != nil {
		t.Fatal(err)
	}
	got, err := bps.AllAliases()
	if err == nil {
		t.Fatal("expected aggregated error for bad underlying")
	}
	if len(got) != 1 || got[0].Type.String() != "test.aliasable_mode" {
		t.Fatalf("expected only the good entry, got %d entries", len(got))
	}
}

// badAliasable returns an underlying primitive that isn't on the allowlist.
type badAliasable struct{}

// astral:blueprint-ignore
func (*badAliasable) ObjectType() string                  { return "test.bad_aliasable" }
func (*badAliasable) UnderlyingPrimitive() string         { return "int32" }
func (*badAliasable) WriteTo(w io.Writer) (int64, error)  { return 0, nil }
func (*badAliasable) ReadFrom(r io.Reader) (int64, error) { return 0, nil }

// TestEncodeDecode_AliasCrossesBlueprints pins the wire path: a value of an Aliasable
// type, encoded via the polymorphic envelope on the originating peer (which has the Go
// type), decodes on a peer that has ONLY the BlueprintAlias as *RuntimeAlias with the
// underlying byte intact.
func TestEncodeDecode_AliasCrossesBlueprints(t *testing.T) {
	// originating peer: typed Go value
	local := NewBlueprints(DefaultBlueprints())
	if err := local.Add(newAliasableMode()); err != nil {
		t.Fatal(err)
	}
	src := aliasableMode(7)
	buf, err := EncodeBytes(&src, WithBlueprints(local))
	if err != nil {
		t.Fatal(err)
	}

	// remote peer: only the schema descriptor (alias), no Go type
	remote := NewBlueprints(DefaultBlueprints())
	if _, err := remote.RegisterAlias(NewBlueprintAlias("test.aliasable_mode", "uint8")); err != nil {
		t.Fatal(err)
	}
	got, _, err := Decode(bytes.NewReader(buf), WithBlueprints(remote))
	if err != nil {
		t.Fatal(err)
	}
	ra, ok := got.(*RuntimeAlias)
	if !ok {
		t.Fatalf("remote decode: want *RuntimeAlias, got %T", got)
	}
	if ra.ObjectType() != "test.aliasable_mode" {
		t.Fatalf("ObjectType drift: %s", ra.ObjectType())
	}
	u, ok := ra.Underlying().(*Uint8)
	if !ok || uint8(*u) != 7 {
		t.Fatalf("payload drift: %#v", ra.Underlying())
	}
}

// TestRuntimeAlias_JSONRoundTrip pins that the JSON shape is the bare underlying value
// (no envelope, no wrapper) and that UnmarshalJSON populates a bound carrier.
func TestRuntimeAlias_JSONRoundTrip(t *testing.T) {
	bps := NewBlueprints(DefaultBlueprints())
	if _, err := bps.RegisterAlias(NewBlueprintAlias("test.alias_json", "uint16")); err != nil {
		t.Fatal(err)
	}
	src := bps.New("test.alias_json").(*RuntimeAlias)
	v := Uint16(42)
	src.value = &v

	data, err := json.Marshal(src)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "42" {
		t.Fatalf("alias JSON should be bare value, got %s", data)
	}

	dst := bps.New("test.alias_json").(*RuntimeAlias)
	if err := json.Unmarshal(data, dst); err != nil {
		t.Fatal(err)
	}
	got, ok := dst.Underlying().(*Uint16)
	if !ok || uint16(*got) != 42 {
		t.Fatalf("JSON round-trip lost value: %#v", dst.Underlying())
	}
}

// TestAliasOf_RejectsNonAliasable pins the AliasOf contract: an Object that doesn't
// implement Aliasable surfaces ErrBlueprintInvalid rather than returning a hollow alias.
func TestAliasOf_RejectsNonAliasable(t *testing.T) {
	b := Bool(false)
	_, err := AliasOf(&b)
	if !errors.Is(err, ErrBlueprintInvalid) {
		t.Fatalf("want ErrBlueprintInvalid for non-Aliasable, got %v", err)
	}
}

// TestAliasOf_RejectsBadUnderlying pins that an Aliasable returning an unallowed primitive
// is caught by AliasOf, not silently accepted.
func TestAliasOf_RejectsBadUnderlying(t *testing.T) {
	_, err := AliasOf(&badAliasable{})
	if !errors.Is(err, ErrBlueprintInvalid) {
		t.Fatalf("want ErrBlueprintInvalid for bad underlying, got %v", err)
	}
}

// TestAliasOf_Aliasable pins the happy path.
func TestAliasOf_Aliasable(t *testing.T) {
	got, err := AliasOf(newAliasableMode())
	if err != nil {
		t.Fatal(err)
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
		t.Fatalf("Register(*Blueprint) failed: %v", err)
	}
	if bps.GetBlueprint("test.register_bp") == nil {
		t.Fatal("Register(*Blueprint) did not store a Blueprint")
	}

	aID, err := bps.Register(NewBlueprintAlias("test.register_alias", "uint16"))
	if err != nil || aID == nil {
		t.Fatalf("Register(*BlueprintAlias) failed: %v", err)
	}
	if bps.GetAlias("test.register_alias") == nil {
		t.Fatal("Register(*BlueprintAlias) did not store an Alias")
	}

	// Anything else must be rejected with ErrBlueprintInvalid.
	b := Bool(false)
	_, err = bps.Register(&b)
	if !errors.Is(err, ErrBlueprintInvalid) {
		t.Fatalf("Register(other) should reject with ErrBlueprintInvalid, got %v", err)
	}
}

func TestGetAlias_CompileTimeProtoNotReturned(t *testing.T) {
	// astral.blueprint.alias is the compile-time prototype for BlueprintAlias itself; it
	// lives in the main map, not the aliases map, so GetAlias must NOT return it.
	if a := DefaultBlueprints().GetAlias("astral.blueprint.alias"); a != nil {
		t.Fatalf("expected nil for compile-time prototype, got %#v", a)
	}
}
