package user

import "github.com/cryptopunkscc/astrald/astral"

var ErrInvalidContract = astral.NewError("invalid contract")
var ErrContractInvalidSignature = astral.NewError("contract invalid signature")
var ErrInvitationDeclined = astral.NewError("invitation declined")
var ErrRequestDeclined = astral.NewError("request declined")
var ErrNodeContractAlreadyExpired = astral.NewError("node contract already expired")
var ErrNodeContractRevocationInvalid = astral.NewError("node contract revocation invalid")
var ErrNodeCannotRevokeContract = astral.NewError("node cannot revoke contract")
var ErrContractNotExists = astral.NewError("node contract not exists")
var ErrContractRevocationNotExists = astral.NewError("node contract revocation not exists")
var ErrNodeContractNotFound = astral.NewError("contract not found")
var ErrContractRevocationNotFound = astral.NewError("contract revocation not found")
