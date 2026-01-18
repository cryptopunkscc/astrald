package crypto

import "errors"

var (
	ErrUnsupportedKeyType = errors.New("unsupported key type")
	ErrInvalidSignature   = errors.New("invalid signature")
	ErrUnsupported        = errors.New("unsupported")
)
