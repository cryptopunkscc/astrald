package user

import "github.com/cryptopunkscc/astrald/astral"

type SwarmJoinRequestPolicy func(requester *astral.Identity) bool
type SwarmInvitePolicy func(invitee *astral.Identity, contract NodeContract) bool
