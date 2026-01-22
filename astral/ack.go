package astral

type Ack struct {
	EmptyObject
}

var _ Object = &Ack{}

// astral

func (Ack) ObjectType() string { return "ack" }

func init() {
	Add(&Ack{})
}
