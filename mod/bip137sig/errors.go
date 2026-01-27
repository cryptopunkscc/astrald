package bip137sig

import (
	"errors"
)

var (
	// ErrInvalidMnemonic indicates the mnemonic is invalid
	ErrInvalidMnemonic          = errors.New("invalid mnemonic")
	ErrInvalidMnemonicWordCount = errors.New("mnemonic must be 12, 15, 18, 21, or 24 words")
	ErrInvalidEntropySize       = errors.New("entropy must be 128-256 bits in 32-bit increments")
	ErrInvalidEntropyLength     = errors.New("invalid entropy length")
	ErrInvalidSeedLength        = errors.New("invalid BIP-39 seed length")
)
