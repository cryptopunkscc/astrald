package astral

import (
	"fmt"
	"reflect"
	"time"
)

// Adapt converts a native Go value to an equivalent astral Object, returns nil if no equivalent exists.
// If v is already an Object, v will be returned. If the argument is nil, &astral.Nil{} will be returned.
func Adapt(v any) Object {
	if v == nil {
		return &Nil{}
	}

	// why: typed-nil pointers (e.g. (*int)(nil)) pass the v == nil check above (interface-nil
	// only fires when both type and value are nil). Without this guard the type-switch below
	// dereferences the nil pointer and panics.
	rv := reflect.ValueOf(v)
	if (rv.Kind() == reflect.Ptr || rv.Kind() == reflect.Interface) && rv.IsNil() {
		return &Nil{}
	}

	if o, ok := v.(Object); ok {
		return o
	}

	switch v := v.(type) {
	case bool:
		return (*Bool)(&v)

	case *bool:
		return (*Bool)(v)

	case int:
		var i = Int64(v)
		return &i

	case *int:
		var i = Int64(*v)
		return &i

	case int8:
		return (*Int8)(&v)

	case *int8:
		return (*Int8)(v)

	case int16:
		return (*Int16)(&v)

	case *int16:
		return (*Int16)(v)

	case int32:
		return (*Int32)(&v)

	case *int32:
		return (*Int32)(v)

	case int64:
		return (*Int64)(&v)

	case *int64:
		return (*Int64)(v)

	case uint:
		var u = Uint64(v)
		return &u

	case *uint:
		var u = Uint64(*v)
		return &u

	case uint8:
		return (*Uint8)(&v)

	case *uint8:
		return (*Uint8)(v)

	case uint16:
		return (*Uint16)(&v)

	case *uint16:
		return (*Uint16)(v)

	case uint32:
		return (*Uint32)(&v)

	case *uint32:
		return (*Uint32)(v)

	case uint64:
		return (*Uint64)(&v)

	case *uint64:
		return (*Uint64)(v)

	case string:
		return (*String32)(&v)

	case *string:
		return (*String32)(v)

	case time.Time:
		return (*Time)(&v)

	case *time.Time:
		return (*Time)(v)

	case time.Duration:
		return (*Duration)(&v)

	case *time.Duration:
		return (*Duration)(v)

	case Bool:
		return &v

	case String8:
		return &v

	case String16:
		return &v

	case String32:
		return &v

	case String64:
		return &v

	case Uint8:
		return &v

	case Uint16:
		return &v

	case Uint32:
		return &v

	case Uint64:
		return &v

	case Int8:
		return &v

	case Int16:
		return &v

	case Int32:
		return &v

	case Int64:
		return &v

	case Nonce:
		return &v

	case Time:
		return &v

	case Duration:
		return &v

	case Bytes8:
		return &v

	case Bytes16:
		return &v

	case Bytes32:
		return &v

	case Bytes64:
		return &v

	case error:
		return NewError(v.Error())

	default:
		return nil
	}
}

// normalize converts a caller-supplied value into the canonical astral value for the given Spec.
func normalize(spec Spec, v any) (Object, error) {
	switch s := spec.(type) {
	case *PrimitiveSpec:
		return normalizePrimitive(s.PrimitiveType.String(), v)

	case *RefSpec:
		if v == nil {
			return nil, fmt.Errorf("RefSpec.Type=%s does not accept nil", s.Type)
		}
		if ro, ok := v.(*RuntimeObject); ok {
			if ro.ObjectType() != s.Type.String() {
				return nil, fmt.Errorf("want %s, got runtime %s", s.Type, ro.ObjectType())
			}
			return ro, nil
		}
		obj, ok := v.(Object)
		if !ok || obj.ObjectType() != s.Type.String() {
			return nil, fmt.Errorf("want %s, got %T", s.Type, v)
		}
		return obj, nil

	case *SliceSpec:
		rs, ok := v.(*RuntimeSlice)
		if !ok {
			return nil, fmt.Errorf("want *RuntimeSlice, got %T", v)
		}
		// why: a *RuntimeSlice over a different element type silently encodes mismatched
		// wire bytes — the caller's `NewRuntimeSlice("A")` for a field whose SliceSpec.Type
		// is "B" must not slip past the Set gate.
		if rs.elemName != s.Type.String() {
			return nil, fmt.Errorf("SliceSpec.Type=%s: got slice of %q", s.Type, rs.elemName)
		}
		return rs, nil
	case *ArraySpec:
		ra, ok := v.(*RuntimeArray)
		if !ok {
			return nil, fmt.Errorf("want *RuntimeArray, got %T", v)
		}
		if uint32(ra.Len()) != uint32(s.Length) {
			return nil, fmt.Errorf("ArraySpec: want length %d, got %d", s.Length, ra.Len())
		}
		if ra.elemName != s.Type.String() {
			return nil, fmt.Errorf("ArraySpec.Type=%s: got array of %q", s.Type, ra.elemName)
		}
		return ra, nil
	case *MapSpec:
		rm, ok := v.(*RuntimeMap)
		if !ok {
			return nil, fmt.Errorf("want *RuntimeMap, got %T", v)
		}
		if rm.valueName != s.ValueType.String() {
			return nil, fmt.Errorf("MapSpec.ValueType=%s: got map of %q", s.ValueType, rm.valueName)
		}
		// why: the carrier's key reflect.Kind is set at construction (resolveKeyType);
		// compare its Kind name against the Spec's KeyType to catch a uint8-Spec field
		// accepting a uint64-keyed carrier and writing a 1-byte key the peer reads as 8.
		gotKey, _ := mapKeyTypeName(rm.ptr.Elem().Type().Key())
		if gotKey != s.KeyType.String() {
			return nil, fmt.Errorf("MapSpec.KeyType=%s: got map with key %s", s.KeyType, gotKey)
		}
		return rm, nil

	case *PtrSpec:
		// why: *Nil is the canonical absent carrier; map Go-nil to it so Get always returns a
		// non-nil zero value (matches the vision contract).
		if v == nil {
			return &Nil{}, nil
		}
		if _, ok := v.(*Nil); ok {
			return &Nil{}, nil
		}

		obj, ok := v.(Object)
		if !ok {
			return nil, fmt.Errorf("PtrSpec.Type=%s: want Object or nil, got %T", s.Type, v)
		}
		if obj.ObjectType() != s.Type.String() {
			return nil, fmt.Errorf("PtrSpec.Type=%s: got %s", s.Type, obj.ObjectType())
		}
		return obj, nil

	case *ObjectSpec:
		if v == nil {
			return nil, fmt.Errorf("ObjectSpec does not accept nil")
		}
		obj, ok := v.(Object)
		if !ok {
			return nil, fmt.Errorf("ObjectSpec: want Object, got %T", v)
		}
		// why: ObjectSpec wire shape is [String8 tag][payload] via Encode/Decode. The runtime
		// carriers report unregistered tags ("slice", "map", "array"), so writing them succeeds
		// but the receiver's Decode resolves nil → ErrBlueprintNotFound. Wrap collections in a
		// named Blueprint or use SliceSpec/MapSpec/ArraySpec directly instead.
		switch obj.(type) {
		case *RuntimeSlice, *RuntimeMap, *RuntimeArray:
			return nil, fmt.Errorf("ObjectSpec: %T has unregistered tag %q; wrap in a named Blueprint", obj, obj.ObjectType())
		}
		return obj, nil
	}

	return nil, fmt.Errorf("unknown spec %T", spec)
}

// normalizePrimitive narrows width-ambiguous Go natives (string, []byte) per spec, then
// delegates everything else to Adapt and verifies the resulting ObjectType matches the spec name.
func normalizePrimitive(name string, v any) (Object, error) {
	switch x := v.(type) {
	case string:
		return narrowString(name, x)
	case *string:
		if x == nil {
			return nil, fmt.Errorf("primitive %s: nil *string", name)
		}

		return narrowString(name, *x)
	case []byte:
		return narrowBytes(name, x)
	}

	obj := Adapt(v)
	if obj == nil || obj.ObjectType() != name {
		return nil, fmt.Errorf("primitive %s: cannot normalize %T", name, v)
	}
	return obj, nil
}

// why: reject oversized inputs at the API boundary so callers get an immediate, informative
// error from ro.Set rather than a generic "data too large" later from WriteTo. WriteTo still
// enforces the same caps as defense-in-depth for direct construction (e.g. String8(rawString)).
func narrowString(name, s string) (Object, error) {
	switch name {
	case "string8":
		if len(s) > 1<<8-1 {
			return nil, fmt.Errorf("primitive string8: length %d exceeds 255", len(s))
		}
		v := String8(s)
		return &v, nil
	case "string16":
		if len(s) > 1<<16-1 {
			return nil, fmt.Errorf("primitive string16: length %d exceeds 65535", len(s))
		}
		v := String16(s)
		return &v, nil
	case "string32":
		if uint64(len(s)) > 1<<32-1 {
			return nil, fmt.Errorf("primitive string32: length %d exceeds 4294967295", len(s))
		}
		v := String32(s)
		return &v, nil
	case "string64":
		v := String64(s)
		return &v, nil
	}
	return nil, fmt.Errorf("primitive %s: cannot accept string", name)
}

func narrowBytes(name string, b []byte) (Object, error) {
	switch name {
	case "bytes8":
		if len(b) > 1<<8-1 {
			return nil, fmt.Errorf("primitive bytes8: length %d exceeds 255", len(b))
		}
		v := Bytes8(b)
		return &v, nil
	case "bytes16":
		if len(b) > 1<<16-1 {
			return nil, fmt.Errorf("primitive bytes16: length %d exceeds 65535", len(b))
		}
		v := Bytes16(b)
		return &v, nil
	case "bytes32":
		if uint64(len(b)) > 1<<32-1 {
			return nil, fmt.Errorf("primitive bytes32: length %d exceeds 4294967295", len(b))
		}
		v := Bytes32(b)
		return &v, nil
	case "bytes64":
		v := Bytes64(b)
		return &v, nil
	}
	return nil, fmt.Errorf("primitive %s: cannot accept []byte", name)
}
