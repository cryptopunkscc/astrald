package shares

import (
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/mod/shares"
)

type dbLocalShare struct {
	Caller  id.Identity
	SetName string
}

func (dbLocalShare) TableName() string { return shares.DBPrefix + "local_shares" }
