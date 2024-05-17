package objects

type Object interface {
	ObjectType() string
}

type Decoder func([]byte) (Object, error)
type Encoder func(object Object) ([]byte, error)
