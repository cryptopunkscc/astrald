package shares

import (
	"github.com/cryptopunkscc/astrald/api"
	"github.com/cryptopunkscc/astrald/components/fid"
)

type Shared interface {
	List(id api.Identity) ([]fid.ID, error)
	Contains(id api.Identity, file fid.ID) (bool, error)
	Add(id api.Identity, file fid.ID) error
	Remove(id api.Identity, file fid.ID) error
}
type Shares interface {
	List() ([]fid.ID, error)
	Contains(file fid.ID) (bool, error)
}
