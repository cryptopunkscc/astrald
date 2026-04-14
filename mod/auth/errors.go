package auth

import "github.com/cryptopunkscc/astrald/astral"

var ErrInvalidContract = astral.NewError("invalid contract")
var ErrContractExpired = astral.NewError("contract expired")
