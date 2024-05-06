package sig

import (
	"context"
	"sync"
)

//TODO: Implemet Flags. Be mindful of memory management:
// - Wait/WaitContext should resuse channels whenever possible
// - Wait/WaitContext should discard channels that are no longer needed

// Flags provides a thread-safe observable flag set
type Flags struct {
	flagStates   map[string]bool
	flagChannels map[string]chan struct{}
	mu           sync.Mutex
}

func NewFlags() *Flags {
	return &Flags{
		flagStates:   make(map[string]bool),
		flagChannels: make(map[string]chan struct{}),
	}
}

// Set sets the provided flags
func (flags *Flags) Set(flag ...string) {
	flags.mu.Lock()
	defer flags.mu.Unlock()

	for _, f := range flag {
		flags.flagStates[f] = true
		if ch, ok := flags.flagChannels[f]; ok {
			close(ch)
			delete(flags.flagChannels, f)
		}
	}
}

// Clear clears the provided flags
func (flags *Flags) Clear(flag ...string) {
	flags.mu.Lock()
	defer flags.mu.Unlock()

	for _, f := range flag {
		flags.flagStates[f] = false
		if ch, ok := flags.flagChannels[f]; ok {
			close(ch)
			delete(flags.flagChannels, f)
		}
	}
}

// IsSet returns true if the flag is up, false otherwise
func (flags *Flags) IsSet(flag string) bool {
	flags.mu.Lock()
	defer flags.mu.Unlock()

	return flags.flagStates[flag]
}

// Flags returns a list of all set flags
func (flags *Flags) Flags() []string {
	flags.mu.Lock()
	defer flags.mu.Unlock()

	var setFlags []string
	for flag, state := range flags.flagStates {
		if state {
			setFlags = append(setFlags, flag)
		}
	}
	return setFlags
}

// Wait returns a channel that will be closed as soon as the flag is in the specified state.
// If the flag is already in the specified state, Wait immediately returns a closed channel.
func (flags *Flags) Wait(flag string, state bool) <-chan struct{} {
	flags.mu.Lock()
	defer flags.mu.Unlock()

	if flags.flagStates[flag] == state {
		ch := make(chan struct{})
		close(ch)
		return ch
	}

	ch, ok := flags.flagChannels[flag]
	if !ok {
		ch = make(chan struct{})
		flags.flagChannels[flag] = ch
	}
	return ch
}

// WaitContext waits until one of the following occurs:
// 1. Context is canceled - WaitContext returns ctx.Err()
// 2. Flag is in the specified state - WaitContext returns nil
// If the flag is already in the specified state when the function is called, it returns nil immediately.
func (flags *Flags) WaitContext(ctx context.Context, flag string, state bool) error {
	flags.mu.Lock()
	if flags.flagStates[flag] == state {
		flags.mu.Unlock()
		return nil
	}

	ch, ok := flags.flagChannels[flag]
	if !ok {
		ch = make(chan struct{})
		flags.flagChannels[flag] = ch
	}
	flags.mu.Unlock()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-ch:
		return nil
	}
}
