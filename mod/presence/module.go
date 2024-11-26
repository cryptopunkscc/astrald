package presence

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/cslq"
	"github.com/cryptopunkscc/astrald/streams"
	"io"
)

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
	Identity *astral.Identity
	Alias    string
	Flags    []string
}

func (p *Presence) ObjectType() string {
	return "mod.presence.presence_info"
}

func (p *Presence) WriteTo(w io.Writer) (n int64, err error) {
	c := streams.NewWriteCounter(w)
	err = cslq.Encode(c, "v [c]c [c][c]c", p.Identity, p.Alias, p.Flags)
	n = c.Total()
	return
}

func (p *Presence) ReadFrom(r io.Reader) (n int64, err error) {
	c := streams.NewReadCounter(r)
	err = cslq.Decode(c, "v [c]c [c][c]c", &p.Identity, &p.Alias, &p.Flags)
	n = c.Total()
	return
}

const (
	DiscoverFlag = "discover"
	SetupFlag    = "setup"
	ActionList   = "mod.presence.list"
)
