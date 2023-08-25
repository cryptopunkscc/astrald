package network

import (
	"errors"
	"time"
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

const MaxPeerLinks = 8
const HandshakeTimeout = 15 * time.Second
