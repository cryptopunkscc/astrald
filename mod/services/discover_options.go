package services

// DiscoverOptions controls snapshot and follow semantics for service discovery streams.
//
// Snapshot:
//
//	When true, the discoverer should emit zero or one initial event representing current state.
//
// Follow:
//
//	When true, the discoverer should continue streaming future changes until ctx cancellation.
type DiscoverOptions struct {
	Snapshot bool
	Follow   bool
}
