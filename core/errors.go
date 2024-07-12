package core

import (
	"errors"
	"fmt"
)

var (
	ErrPeerUnlinked          = errors.New("peer unlinked")
	ErrPeerLinkLimitExceeded = errors.New("link limit exceeded")
	ErrDuplicateLink         = errors.New("duplicate link")
	ErrLinkNotFound          = errors.New("not found")
	ErrNotRunning            = errors.New("not running")
	ErrIdentityMismatch      = errors.New("local identity mismatch")
	ErrLinkIsNil             = errors.New("link is nil")
)

type ErrModuleUnavailable struct {
	Name string
}

func (err ErrModuleUnavailable) Error() string {
	return fmt.Sprintf("module %s unavailable", err.Name)
}

func ModuleUnavailable(name string) ErrModuleUnavailable {
	return ErrModuleUnavailable{Name: name}
}

func (ErrModuleUnavailable) Is(other error) bool {
	_, ok := other.(*ErrModuleUnavailable)
	return ok
}
