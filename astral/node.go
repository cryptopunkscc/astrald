package astral

// Node defines the basic interface of an astral node
type Node interface {
	Router
	Identity() *Identity
}
