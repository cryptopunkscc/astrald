package nat

import "errors"

var ErrDuplicatePair = errors.New("duplicate pair")
var ErrPairNotExists = errors.New("pair not exists")
var ErrPairBusy = errors.New("pair is busy")
var ErrPairCantLock = errors.New("pair can't lock")
var ErrNoSuitableIP = errors.New("no suitable IPv4 address found")
