package astral

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"reflect"
	"strconv"
)

var _ Object = (*RuntimeMap)(nil)

// RuntimeMap is the carrier for MapSpec fields in runtime Blueprints. It owns a typed native
// map (heterogeneous map[K]Object or homogeneous map[K]*T) and delegates encoding to the
// reflective mapValue codec for byte-identical parity with Objectify on a native map field.
//
// When valueName is a runtime Blueprint, ReadFrom takes a slow path that constructs each
// value via New(valueName); see RuntimeSlice for the same rationale.
type RuntimeMap struct {
	ptr       reflect.Value // *map[keyType]elemType, always addressable
	valueName string        // MapSpec.ValueType — empty means heterogeneous
}

// astral:blueprint-ignore
func (*RuntimeMap) ObjectType() string { return "map" }

// NewRuntimeMap returns a RuntimeMap whose key and element types are determined from a
// MapSpec. An empty valueType means heterogeneous (element type = Object interface).
// Returns an error if keyType is unsupported or if valueType is non-empty and unregistered.
// Element resolution uses defaultBlueprints; the codec path uses newRuntimeMapWith
// internally for per-call registries.
func NewRuntimeMap(keyType, valueType string) (*RuntimeMap, error) {
	return newRuntimeMapWith(defaultBlueprints, keyType, valueType)
}

// NewRuntimeMapWith is the bps-aware constructor for callers that hold a *Blueprints (e.g.
// a per-call registry built around DefaultBlueprints). Same shape as NewRuntimeMap; value
// resolution consults bps instead of the package default. Use when populating a RuntimeObject
// field whose MapSpec.ValueType names a Blueprint registered only in bps.
func NewRuntimeMapWith(bps *Blueprints, keyType, valueType string) (*RuntimeMap, error) {
	return newRuntimeMapWith(bps, keyType, valueType)
}

// newRuntimeMapWith is the internal bps-aware constructor used by readField when decoding
// through a custom registry (WithBlueprints).
func newRuntimeMapWith(bps *Blueprints, keyType, valueType string) (*RuntimeMap, error) {
	kt, err := resolveKeyType(keyType)
	if err != nil {
		return nil, err
	}
	et, err := resolveElemType(bps, valueType)
	if err != nil {
		return nil, err
	}
	return &RuntimeMap{
		ptr:       reflect.New(reflect.MapOf(kt, et)),
		valueName: valueType,
	}, nil
}

// resolveKeyType maps a MapSpec.KeyType name to its Go reflect.Type. The supported set mirrors
// mapKeyAllowlist exactly: "string16" → string, "uintN" → uintN for N ∈ {8,16,32,64}.
func resolveKeyType(name string) (reflect.Type, error) {
	switch name {
	case "string16":
		return reflect.TypeOf(""), nil
	case "uint8":
		return reflect.TypeOf(uint8(0)), nil
	case "uint16":
		return reflect.TypeOf(uint16(0)), nil
	case "uint32":
		return reflect.TypeOf(uint32(0)), nil
	case "uint64":
		return reflect.TypeOf(uint64(0)), nil
	}
	return nil, fmt.Errorf("runtime_map: unsupported key type %q", name)
}

func (m *RuntimeMap) WriteTo(w io.Writer) (int64, error) {
	return mapValue{Value: m.ptr.Elem()}.WriteTo(w)
}

func (m *RuntimeMap) ReadFrom(r io.Reader) (int64, error) {
	bps := blueprintsFromReader(r)
	if !isRuntimeBlueprintType(bps, m.valueName) {
		return mapValue{Value: m.ptr.Elem()}.ReadFrom(r)
	}
	return m.readRuntimeBlueprintEntries(r, bps)
}

func (m *RuntimeMap) readRuntimeBlueprintEntries(r io.Reader, bps *Blueprints) (int64, error) {
	val := m.ptr.Elem()
	keyWidth, ok := supportedMapKey(val.Type().Key().Kind())
	if !ok {
		return 0, fmt.Errorf("runtime_map: unsupported key kind %s", val.Type().Key().Kind())
	}

	var l uint32
	err := binary.Read(r, ByteOrder, &l)
	if err != nil {
		return 0, err
	}
	var n int64 = 4

	if l == 0 {
		val.SetZero()
		return n, nil
	}

	val.Set(reflect.MakeMapWithSize(val.Type(), int(l)))
	for i := uint32(0); i < l; i++ {
		key := reflect.New(val.Type().Key()).Elem()
		km, err := readMapKey(r, key, keyWidth)
		n += km
		if err != nil {
			return n, err
		}

		elem, vm, err := readRuntimeBlueprintMapValue(r, bps, m.valueName)
		n += vm
		if err != nil {
			return n, err
		}
		if elem == nil {
			// why: nil-flag was 0; insert a zero-valued *RuntimeObject pointer to mirror
			// what reflect.MakeMap + SetMapIndex would produce on the heterogeneous path.
			val.SetMapIndex(key, reflect.Zero(val.Type().Elem()))
			continue
		}
		val.SetMapIndex(key, reflect.ValueOf(elem))
	}
	return n, nil
}

func (m *RuntimeMap) MarshalJSON() ([]byte, error) {
	return mapValue{Value: m.ptr.Elem()}.MarshalJSON()
}

// UnmarshalJSON: see RuntimeSlice.UnmarshalJSON note — JSON path uses defaultBlueprints only.
func (m *RuntimeMap) UnmarshalJSON(data []byte) error {
	if !isRuntimeBlueprintType(defaultBlueprints, m.valueName) {
		return mapValue{Value: m.ptr.Elem()}.UnmarshalJSON(data)
	}
	val := m.ptr.Elem()
	if bytes.Equal(data, jsonNull) {
		val.SetZero()
		return nil
	}
	keyType := val.Type().Key()
	if _, ok := supportedMapKey(keyType.Kind()); !ok {
		return fmt.Errorf("runtime_map: unsupported key kind %s", keyType.Kind())
	}
	var raw map[string]json.RawMessage
	err := json.Unmarshal(data, &raw)
	if err != nil {
		return err
	}
	val.Set(reflect.MakeMap(val.Type()))
	for k, v := range raw {
		mapKey := reflect.New(keyType).Elem()
		switch keyType.Kind() {
		case reflect.String:
			mapKey.SetString(k)
		default:
			// why: parse at the actual width so overflowing keys error out instead of
			// silently truncating via reflect.SetUint. Mirrors Set→convertMapKey's range
			// check at runtime_map.go:263.
			bitSize := int(keyType.Size()) * 8
			u, perr := strconv.ParseUint(k, 10, bitSize)
			if perr != nil {
				return fmt.Errorf("runtime_map: parse uint key %q: %w", k, perr)
			}
			mapKey.SetUint(u)
		}
		slot := reflect.New(val.Type().Elem()).Elem()
		err := unmarshalRuntimeBlueprintPtr(v, m.valueName, slot)
		if err != nil {
			return err
		}
		val.SetMapIndex(mapKey, slot)
	}
	return nil
}

// Set assigns value under key. key must be string for string-keyed maps or uint64 for uintN-keyed
// maps; any other Go type is rejected. For narrow uint widths the key is range-checked. value's
// runtime type must be assignable to the carrier's element type.
func (m *RuntimeMap) Set(key any, value Object) error {
	keyT := m.ptr.Elem().Type().Key()
	kv, err := convertMapKey(keyT, key)
	if err != nil {
		return err
	}
	elemT := m.ptr.Elem().Type().Elem()
	rv := reflect.ValueOf(value)
	if !rv.IsValid() || !rv.Type().AssignableTo(elemT) {
		return fmt.Errorf("runtime_map: want %s, got %T", elemT, value)
	}
	if m.ptr.Elem().IsNil() {
		m.ptr.Elem().Set(reflect.MakeMap(m.ptr.Elem().Type()))
	}
	m.ptr.Elem().SetMapIndex(kv, rv)
	return nil
}

// Get returns the value stored under key. key follows the same kind contract as Set.
func (m *RuntimeMap) Get(key any) (Object, bool) {
	if m.ptr.Elem().IsNil() {
		return nil, false
	}
	keyT := m.ptr.Elem().Type().Key()
	kv, err := convertMapKey(keyT, key)
	if err != nil {
		return nil, false
	}
	v := m.ptr.Elem().MapIndex(kv)
	if !v.IsValid() {
		return nil, false
	}
	return v.Interface().(Object), true
}

func (m *RuntimeMap) Len() int {
	return m.ptr.Elem().Len()
}

// Each iterates over entries in unspecified order. The key is passed as string for string-keyed
// maps or uint64 for uintN-keyed maps. Stop iteration by returning a non-nil error.
func (m *RuntimeMap) Each(fn func(key any, value Object) error) error {
	if m.ptr.Elem().IsNil() {
		return nil
	}
	iter := m.ptr.Elem().MapRange()
	for iter.Next() {
		var k any
		if iter.Key().Kind() == reflect.String {
			k = iter.Key().String()
		} else {
			k = iter.Key().Uint()
		}
		if err := fn(k, iter.Value().Interface().(Object)); err != nil {
			return err
		}
	}
	return nil
}

// convertMapKey narrows a caller-supplied key into the configured key type. Strict contract:
// string keys accept only string; uint keys accept only uint64. Narrow uint widths are
// range-checked.
func convertMapKey(keyT reflect.Type, key any) (reflect.Value, error) {
	kv := reflect.New(keyT).Elem()
	switch keyT.Kind() {
	case reflect.String:
		s, ok := key.(string)
		if !ok {
			return reflect.Value{}, fmt.Errorf("runtime_map: want string key, got %T", key)
		}
		kv.SetString(s)
	case reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		u, ok := key.(uint64)
		if !ok {
			return reflect.Value{}, fmt.Errorf("runtime_map: want uint64 key, got %T", key)
		}
		if keyT.Kind() != reflect.Uint64 && u>>(uint(keyT.Size())*8) != 0 {
			return reflect.Value{}, fmt.Errorf("runtime_map: key %d overflows width %d", u, keyT.Size())
		}
		kv.SetUint(u)
	default:
		return reflect.Value{}, fmt.Errorf("runtime_map: unsupported key kind %s", keyT.Kind())
	}
	return kv, nil
}
