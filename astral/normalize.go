package astral

import (
	"fmt"
	"time"
)

// Adapt converts a native Go value to an equivalent astral Object, returns nil if no equivalent exists.
// If v is already an Object, v will be returned. If the argument is nil, &astral.Nil{} will be returned.
func Adapt(v any) Object {
	if v == nil {
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
func normalize(spec Object, v any) (Object, error) {
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
		rs, ok := v.(*runtimeSlice)
		if !ok {
			return nil, fmt.Errorf("want *runtimeSlice, got %T", v)
		}
		return rs, nil
	case *MapSpec:
		rm, ok := v.(*runtimeMap)
		if !ok {
			return nil, fmt.Errorf("want *runtimeMap, got %T", v)
		}
		return rm, nil

	case *PtrSpec:
		if v == nil {
			return nil, nil
		}

		obj, ok := v.(Object)
		if !ok {
			return nil, fmt.Errorf("PtrSpec.Type=%s: want Object or nil, got %T", s.Type, v)
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

func narrowString(name, s string) (Object, error) {
	switch name {
	case "string8":
		v := String8(s)
		return &v, nil
	case "string16":
		v := String16(s)
		return &v, nil
	case "string32":
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
		v := Bytes8(b)
		return &v, nil
	case "bytes16":
		v := Bytes16(b)
		return &v, nil
	case "bytes32":
		v := Bytes32(b)
		return &v, nil
	case "bytes64":
		v := Bytes64(b)
		return &v, nil
	}
	return nil, fmt.Errorf("primitive %s: cannot accept []byte", name)
}
