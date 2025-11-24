package user

import "github.com/cryptopunkscc/astrald/astral"

var ErrInvalidContract = astral.NewError("invalid contract")
var ErrContractInvalidSignature = astral.NewError("contract invalid signature")
var ErrInvitationDeclined = astral.NewError("invitation declined")
var ErrRequestDeclined = astral.NewError("request declined")
var ErrNodeContractAlreadyExpired = astral.NewError("node contract already expired")
var ErrNodeContractRevocationInvalid = astral.NewError("node contract revocation invalid")
var ErrNodeContractRevocationForExpiredContract = astral.NewError("node contract revocation is for expired contract")
