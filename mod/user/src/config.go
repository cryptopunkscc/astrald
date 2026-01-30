package user

import (
	"time"

	"github.com/cryptopunkscc/astrald/mod/tree"
	"github.com/cryptopunkscc/astrald/mod/user"
)

const (
	minimalRevocationLength = 7 * 24 * time.Hour
	minimalContractLength   = time.Hour
	defaultContractValidity = 365 * 24 * time.Hour
)

type Config struct {
	ActiveContract tree.Value[*user.SignedNodeContract]
}
