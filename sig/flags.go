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
	flagMap map[string]bool
	waitMap map[string][]chan struct{}
	mtx     sync.Mutex
}

func NewFlags() *Flags {
	flag := &Flags{
		flagMap: make(map[string]bool),
		waitMap: make(map[string][]chan struct{}),
	}

	return flag
}

// Set sets the provided flags
func (flags *Flags) Set(flag ...string) {
	flags.mtx.Lock()
	defer flags.mtx.Unlock()

	for _, f := range flag {
		flags.flagMap[f] = true
	}

	for _, f := range flag {
		for _, ch := range flags.waitMap[f] {
			close(ch)
		}
		flags.waitMap[f] = nil
	}
}

// Clear clears the provided flags
func (flags *Flags) Clear(flag ...string) {
	flags.mtx.Lock()
	defer flags.mtx.Unlock()

	for _, f := range flag {
		flags.flagMap[f] = false
	}

	for _, f := range flag {
		for _, ch := range flags.waitMap[f] {
			close(ch)
		}
		flags.waitMap[f] = nil
	}
}

// IsSet returns true if the flag is up, false otherwise
func (flags *Flags) IsSet(flag string) bool {
	flags.mtx.Lock()
	defer flags.mtx.Unlock()
	return flags.flagMap[flag]
}

// Flags returns a list of all set flags
func (flags *Flags) Flags() []string {
	flags.mtx.Lock()
	defer flags.mtx.Unlock()

	var setFlags []string
	for flag, isSet := range flags.flagMap {
		if isSet {
			setFlags = append(setFlags, flag)
		}
	}

	return setFlags
}

// Wait returns a channel that will be closed as soon as the flag is in the specified state.
// If the flag is already in the specified state, Wait immediately returns a closed channel.
func (flags *Flags) Wait(flag string, state bool) <-chan struct{} {
	flags.mtx.Lock()
	defer flags.mtx.Unlock()

	if flags.flagMap[flag] == state {
		ch := make(chan struct{})
		close(ch)
		return ch
	}

	ch := make(chan struct{})
	flags.waitMap[flag] = append(flags.waitMap[flag], ch)

	return ch
}

// WaitContext waits until one of the following occurs:
// 1. Context is canceled - WaitContext returns ctx.Err()
// 2. Flag is in the specified state - WaitContext returns nil
// If the flag is already in the specified state when the function is called, it returns nil immediately.
func (flags *Flags) WaitContext(ctx context.Context, flag string, state bool) error {
	flags.mtx.Lock()
	defer flags.mtx.Unlock()

	if flags.flagMap[flag] == state {
		return nil
	}

	ch := make(chan struct{})
	flags.waitMap[flag] = append(flags.waitMap[flag], ch)

	flags.mtx.Unlock()

	select {
	case <-ch:
		return nil

	case <-ctx.Done():

		return ctx.Err()
	}
}
