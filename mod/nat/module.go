package nat

const ModuleName = "nat"

// Module defines the NAT traversal module public API.
// Keep minimal for now; NAT orchestration will be implemented in src/.
type Module interface {
	// Future entry points for traversal coordination will be added here.
}

// NOTE: in objects this lies in src/config.go
// but i believe that other modules could want to refer to it (
// and they can only import this package

const (
	MethodStartNatTraversal = "nat.start_traversal"
)
