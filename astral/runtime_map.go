package astral

import (
	"fmt"
	"io"
	"reflect"
)

var _ Object = (*runtimeMap)(nil)

// runtimeMap is the carrier for MapSpec fields in runtime Blueprints. It owns a typed native
// map (heterogeneous map[K]Object or homogeneous map[K]*T) and delegates encoding to the
// reflective mapValue codec for byte-identical parity with Objectify on a native map field.
type runtimeMap struct {
	ptr reflect.Value // *map[keyType]elemType, always addressable
}

// astral:blueprint-ignore
func (*runtimeMap) ObjectType() string { return "map" }

// newRuntimeMap returns a runtimeMap whose key and element types are determined from a
// MapSpec. An empty valueType means heterogeneous (element type = Object interface).
// Returns an error if keyType is unsupported or if valueType is non-empty and unregistered.
func newRuntimeMap(keyType, valueType string) (*runtimeMap, error) {
	kt, err := resolveKeyType(keyType)
	if err != nil {
		return nil, err
	}
	et, err := resolveElemType(valueType)
	if err != nil {
		return nil, err
	}
	return &runtimeMap{ptr: reflect.New(reflect.MapOf(kt, et))}, nil
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

func (m *runtimeMap) WriteTo(w io.Writer) (int64, error) {
	return mapValue{Value: m.ptr.Elem()}.WriteTo(w)
}

func (m *runtimeMap) ReadFrom(r io.Reader) (int64, error) {
	return mapValue{Value: m.ptr.Elem()}.ReadFrom(r)
}

func (m *runtimeMap) MarshalJSON() ([]byte, error) {
	return mapValue{Value: m.ptr.Elem()}.MarshalJSON()
}

func (m *runtimeMap) UnmarshalJSON(data []byte) error {
	return mapValue{Value: m.ptr.Elem()}.UnmarshalJSON(data)
}

// Set assigns value under key. key must be string for string-keyed maps or uint64 for uintN-keyed
// maps; any other Go type is rejected. For narrow uint widths the key is range-checked. value's
// runtime type must be assignable to the carrier's element type.
func (m *runtimeMap) Set(key any, value Object) error {
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
func (m *runtimeMap) Get(key any) (Object, bool) {
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

func (m *runtimeMap) Len() int {
	return m.ptr.Elem().Len()
}

// Each iterates over entries in unspecified order. The key is passed as string for string-keyed
// maps or uint64 for uintN-keyed maps. Stop iteration by returning a non-nil error.
func (m *runtimeMap) Each(fn func(key any, value Object) error) error {
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
