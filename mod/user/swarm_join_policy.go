package user

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/mod/auth"
)

type SwarmJoinRequestPolicy func(requester *astral.Identity) bool
type SwarmInvitePolicy func(invitee *astral.Identity, contract *auth.Contract) bool
