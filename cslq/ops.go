package cslq

// OpUint8 codes basic uint8 type
type OpUint8 struct{}

// OpUint16 codes basic uint16 type
type OpUint16 struct{}

// OpUint32 codes basic uint32 type
type OpUint32 struct{}

// OpUint64 codes basic uint64 type
type OpUint64 struct{}

// OpInterface codes value using Marshaler or Unmarshaler interfaces.
type OpInterface struct{}

// OpStruct codes a structure using reflection
type OpStruct []interface{}

// OpArray codes an array of basic and complex types
type OpArray struct {
	FixedLen int
	LenOp    interface{}
	ElemOp   interface{}
}
