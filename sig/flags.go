package sig

import "context"

//TODO: Implemet Flags. Be mindful of memory management:
// - Wait/WaitContext should resuse channels whenever possible
// - Wait/WaitContext should discard channels that are no longer needed

// Flags provides a thread-safe observable flag set
type Flags struct {
}

func NewFlags() *Flags {
	return &Flags{}
}

// Set sets the provided flags
func (flags *Flags) Set(flag ...string) {
	//TODO implement me
	panic("implement me")
}

// Clear clears the provided flags
func (flags *Flags) Clear(flag ...string) {
	//TODO implement me
	panic("implement me")
}

// IsSet returns true if the flag is up, false otherwise
func (flags *Flags) IsSet(flag string) bool {
	//TODO implement me
	panic("implement me")
}

// Flags returns a list of all set flags
func (flags *Flags) Flags() []string {
	//TODO implement me
	panic("implement me")
}

// Wait returns a channel that will be closed as soon as the flag is in the specified state.
// If the flag is already in the specified state, Wait immediately returns a closed channel.
func (flags *Flags) Wait(flag string, state bool) <-chan struct{} {
	//TODO implement me
	panic("implement me")
}

// WaitContext waits until one of the following occurs:
// 1. Context is canceled - WaitContext returns ctx.Err()
// 2. Flag is in the specified state - WaitContext returns nil
// If the flag is already in the specified state when the function is called, it returns nil immediately.
func (flags *Flags) WaitContext(ctx context.Context, flag string, state bool) error {
	//TODO implement me
	panic("implement me")
}
