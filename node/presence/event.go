package presence

import (
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/infra"
	"github.com/cryptopunkscc/astrald/node/event"
)

var _ event.Eventer = &Event{}

const (
	EventIdentityPresent = "presence.identity_present"
	EventIdentityGone    = "presence.identity_gone"
)

type Event struct {
	identity id.Identity
	event    string
	addr     infra.Addr
}

func (e Event) Addr() infra.Addr {
	return e.addr
}

func (e Event) Identity() id.Identity {
	return e.identity
}

func (e Event) Event() string {
	return e.event
}
