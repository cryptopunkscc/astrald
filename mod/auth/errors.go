package auth

import "errors"

var ErrInvalidContract = errors.New("invalid contract")
var ErrContractExpired = errors.New("contract expired")
var ErrAlreadySigned = errors.New("already signed")
