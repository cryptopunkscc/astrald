package peers

import (
	"errors"
	_log "github.com/cryptopunkscc/astrald/log"
	"time"
)

var log = _log.Tag("peers")

var (
	ErrPeerUnlinked          = errors.New("peer unlinked")
	ErrPeerLinkLimitExceeded = errors.New("link limit exceeded")
	ErrDuplicateLink         = errors.New("duplicate link")
	ErrLinkNotFound          = errors.New("not found")
	ErrNotRunning            = errors.New("not running")
	ErrIdentityMismatch      = errors.New("local identity mismatch")
)

const MaxPeerLinks = 8
const HandshakeTimeout = 15 * time.Second
