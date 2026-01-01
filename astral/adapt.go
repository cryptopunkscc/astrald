package astral

import (
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

	case error:
		return NewError(v.Error())

	default:
		return nil
	}
}
