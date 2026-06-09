package astral

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"reflect"
)

// Helpers shared by RuntimeSlice / RuntimeArray / RuntimeMap to decode an element whose
// type is a runtime Blueprint. The generic reflective codec allocates *RuntimeObject via
// reflect.New, producing an unbound carrier that silently no-ops on ReadFrom. These helpers
// construct elements via New(typeName) so each carries its schema binding.

// isRuntimeBlueprintType reports whether typeName names a registered runtime Blueprint
// (as opposed to a compile-time prototype or an empty/heterogeneous slot). bps is the
// registry to consult; pass defaultBlueprints when no per-call registry is in play.
func isRuntimeBlueprintType(bps *Blueprints, typeName string) bool {
	return typeName != "" && bps.GetBlueprint(typeName) != nil
}

// readRuntimeBlueprintPtr decodes a single *RuntimeObject slot from r with the wire shape
// that ptrValue would emit: [uint8 nil-flag][element payload if present]. dst must be a
// settable *RuntimeObject slot. bps resolves typeName → element carrier. Returns total
// bytes consumed.
func readRuntimeBlueprintPtr(r io.Reader, bps *Blueprints, typeName string, dst reflect.Value) (int64, error) {
	var nilFlag uint8
	err := binary.Read(r, ByteOrder, &nilFlag)
	if err != nil {
		return 0, err
	}
	if nilFlag == 0 {
		return 1, nil
	}
	if nilFlag != 1 {
		return 1, fmt.Errorf("runtime_blueprint_codec: invalid nil flag %d", nilFlag)
	}
	elem, ok := bps.New(typeName).(*RuntimeObject)
	if !ok {
		return 1, fmt.Errorf("%w: %s", ErrBlueprintNotFound, typeName)
	}
	m, err := elem.ReadFrom(r)
	if err != nil {
		return 1 + m, err
	}
	dst.Set(reflect.ValueOf(elem))
	return 1 + m, nil
}

// readRuntimeBlueprintMapValue decodes a single RuntimeObject value (no nil-flag, no pointer)
// into a map value slot whose element type is *RuntimeObject. The map's wire shape carries
// the value via the generic objectify path which would write the *RuntimeObject directly
// through ptrValue — so the encoded value carries the same nil-flag as a slice/array slot.
// See readRuntimeBlueprintPtr for the nil-flag handling. bps resolves typeName → carrier.
func readRuntimeBlueprintMapValue(r io.Reader, bps *Blueprints, typeName string) (Object, int64, error) {
	var nilFlag uint8
	err := binary.Read(r, ByteOrder, &nilFlag)
	if err != nil {
		return nil, 0, err
	}
	if nilFlag == 0 {
		return nil, 1, nil
	}
	if nilFlag != 1 {
		return nil, 1, fmt.Errorf("runtime_blueprint_codec: invalid nil flag %d", nilFlag)
	}
	elem, ok := bps.New(typeName).(*RuntimeObject)
	if !ok {
		return nil, 1, fmt.Errorf("%w: %s", ErrBlueprintNotFound, typeName)
	}
	m, err := elem.ReadFrom(r)
	if err != nil {
		return nil, 1 + m, err
	}
	return elem, 1 + m, nil
}

// blueprintsFromReader returns the per-call registry threaded through the wrapper, or
// defaultBlueprints when r isn't a *objectReader. Centralizes the type-assert pattern used
// by Runtime{Slice,Array,Map}.ReadFrom to recover bps for nested element resolution.
func blueprintsFromReader(r io.Reader) *Blueprints {
	if or, ok := r.(*objectReader); ok {
		return or.resolve()
	}
	return defaultBlueprints
}

// unmarshalRuntimeBlueprintPtr decodes a JSON value into a *RuntimeObject slot. Null
// resolves to a nil pointer; otherwise a fresh bound *RuntimeObject is constructed via
// New(typeName) before json.Unmarshal runs.
func unmarshalRuntimeBlueprintPtr(raw json.RawMessage, typeName string, dst reflect.Value) error {
	if bytes.Equal(bytes.TrimSpace(raw), jsonNull) {
		dst.Set(reflect.Zero(dst.Type()))
		return nil
	}
	elem, ok := New(typeName).(*RuntimeObject)
	if !ok {
		return fmt.Errorf("%w: %s", ErrBlueprintNotFound, typeName)
	}
	err := json.Unmarshal(raw, elem)
	if err != nil {
		return err
	}
	dst.Set(reflect.ValueOf(elem))
	return nil
}
