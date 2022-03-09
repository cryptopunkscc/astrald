package cslq

// uint tokens

type TokenUint8 struct{}
type TokenUint16 struct{}
type TokenUint32 struct{}
type TokenUint64 struct{}

func (TokenUint8) Compile(_ TokenReader) (Op, error) {
	return OpUint8{}, nil
}

func (TokenUint16) Compile(_ TokenReader) (Op, error) {
	return OpUint16{}, nil
}

func (TokenUint32) Compile(_ TokenReader) (Op, error) {
	return OpUint32{}, nil
}

func (TokenUint64) Compile(_ TokenReader) (Op, error) {
	return OpUint64{}, nil
}
