package apphost

import "github.com/cryptopunkscc/astrald/astral"

var ErrNilContract = astral.NewError("nil contract")
var ErrHostSigMissing = astral.NewError("host signature is missing")
var ErrAppSigMissing = astral.NewError("app signature is missing")
var ErrContractHashFailed = astral.NewError("cannot compute contract hash")
