package bip137sig

import (
	"errors"
)

var (
	// ErrInvalidMnemonic indicates the mnemonic is invalid
	ErrInvalidMnemonic      = errors.New("invalid mnemonic")
	ErrInvalidEntropyLength = errors.New("invalid entropy length")
	ErrInvalidSeedLength    = errors.New("invalid BIP-39 seed length")
)
