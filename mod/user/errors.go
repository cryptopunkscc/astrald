package user

import "github.com/cryptopunkscc/astrald/astral"

var ErrInvalidContract = astral.NewError("invalid contract")
var ErrContractInvalidSignature = astral.NewError("contract invalid signature")
var ErrInvitationDeclined = astral.NewError("invitation declined")
var ErrRequestDeclined = astral.NewError("request declined")
var ErrContractNotExists = astral.NewError("contract not exists")
