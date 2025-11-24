package user

import "github.com/cryptopunkscc/astrald/astral"

var ErrInvalidContract = astral.NewError("invalid contract")
var ErrInvalidSignature = astral.NewError("invalid signature")
var ErrInvitationDeclined = astral.NewError("invitation declined")
var ErrRequestDeclined = astral.NewError("request declined")
