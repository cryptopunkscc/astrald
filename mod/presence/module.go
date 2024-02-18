package presence

import "github.com/cryptopunkscc/astrald/auth/id"

const ModuleName = "presence"

type Module interface {
	Broadcast(flags ...string) error // broadcast node's presence with provided flags
	List() []*Presence
	SetVisible(bool) error
	Visible() bool

	AddHookAdOut(AdOutHook) error
	RemoveHookAdOut(AdOutHook) error
}

type PendingAd interface {
	AddFlag(string)
}

type AdOutHook interface {
	OnPendingAd(PendingAd)
}

type Presence struct {
	Identity id.Identity
	Alias    string
	Flags    []string
}

const (
	DiscoverFlag = "discover"
	SetupFlag    = "setup"
	ScanAction   = "mod.presence.scan"
)
