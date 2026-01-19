package crypto

import "errors"

var (
	ErrUnsupportedKeyType = errors.New("unsupported key type")
	ErrUnsupportedScheme  = errors.New("unsupported scheme")
	ErrInvalidSignature   = errors.New("invalid signature")
	ErrUnsupported        = errors.New("unsupported")
)
