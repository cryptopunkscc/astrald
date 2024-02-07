package presence

import "github.com/cryptopunkscc/astrald/auth/id"

const ModuleName = "presence"

type Module interface {
	List() []*Presence
	SetVisible(bool) error
	Visible() bool
	SetFlags(...string) error
	ClearFlags(...string) error
	Flags() []string
}

type Presence struct {
	Identity id.Identity
	Alias    string
	Flags    []string
}

const (
	DiscoverFlag = "discover"
	PairingFlag  = "pairing"
)
