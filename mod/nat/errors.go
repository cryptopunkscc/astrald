package nat

import "errors"

var ErrDuplicatePair = errors.New("duplicate pair")
var ErrPairNotExists = errors.New("pair not exists")
var ErrPairBusy = errors.New("pair is busy")
var ErrMissingNonce = errors.New("missing nonce")
var ErrPairNotLocked = errors.New("pair not locked")
