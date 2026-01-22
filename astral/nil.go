package astral

// Nil is a pseudo-object that represents the nil value.
type Nil struct {
	EmptyObject
}

var _ Object = &Nil{}

func (Nil) ObjectType() string { return "nil" }

func init() {
	Add(&Nil{})
}
