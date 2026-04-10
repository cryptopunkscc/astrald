package user

import (
	"time"

	"github.com/cryptopunkscc/astrald/mod/auth"
	"github.com/cryptopunkscc/astrald/mod/tree"
)

const (
	minimalContractLength   = time.Hour
	defaultContractValidity = 365 * 24 * time.Hour
)

type Config struct {
	ActiveContract tree.Value[*auth.SignedContract]
}
