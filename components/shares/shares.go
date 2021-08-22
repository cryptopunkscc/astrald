package shares

import (
	"github.com/cryptopunkscc/astrald/api"
	"github.com/cryptopunkscc/astrald/components/fid"
)

type Shares interface {
	Add(id api.Identity, file fid.ID) error
	Remove(id api.Identity, file fid.ID) error
	List(id api.Identity) ([]fid.ID, error)
	Contains(id api.Identity, file fid.ID) (bool, error)
}
