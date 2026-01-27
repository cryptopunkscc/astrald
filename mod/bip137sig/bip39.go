package bip137sig

import (
	"crypto/rand"
	"crypto/sha256"
	"crypto/sha512"
	"errors"
	"fmt"
	"slices"
	"strings"

	"golang.org/x/crypto/pbkdf2"
)

const (
	MinEntropyBits   = 128
	MaxEntropyBits   = 256
	EntropyStepBits  = 32
	MinEntropyBytes  = MinEntropyBits / 8
	MaxEntropyBytes  = MaxEntropyBits / 8
	EntropyStepBytes = EntropyStepBits / 8
	SeedLengthBytes  = 64

	DefaultEntropyBits = MinEntropyBits
)

// reference https://github.com/bitcoin/bips/blob/master/bip-0039.mediawiki

// EntropyToMnemonic converts entropy bytes to mnemonic words.
// Entropy must be 16, 20, 24, 28, or 32 bytes (128-256 bits in 32-bit increments).
func EntropyToMnemonic(entropy Entropy) ([]string, error) {
	entropyBits := len(entropy) * 8
	if entropyBits < MinEntropyBits || entropyBits > MaxEntropyBits || entropyBits%EntropyStepBits != 0 {
		return nil, errors.New("entropy must be 128-256 bits in 32-bit increments")
	}

	// CS = ENT / 32
	checksumBits := entropyBits / 32
	hash := sha256.Sum256(entropy)

	// Combine entropy + checksum bits
	totalBits := entropyBits + checksumBits
	bits := make([]bool, totalBits)

	for i := 0; i < entropyBits; i++ {
		bits[i] = (entropy[i/8] & (1 << (7 - (i % 8)))) != 0
	}
	for i := 0; i < checksumBits; i++ {
		bits[entropyBits+i] = (hash[0] & (1 << (7 - i))) != 0
	}

	// Convert to words (11 bits per word)
	wordCount := totalBits / 11
	words := make([]string, wordCount)
	for i := 0; i < wordCount; i++ {
		idx := 0
		for j := 0; j < 11; j++ {
			if bits[i*11+j] {
				idx |= 1 << (10 - j)
			}
		}
		words[i] = wordlist[idx]
	}

	return words, nil
}

// MnemonicToEntropy extracts entropy bytes from mnemonic words.
// Also validates checksum. Returns error if invalid.
func MnemonicToEntropy(words []string) (Entropy, error) {
	wordCount := len(words)
	if wordCount != 12 && wordCount != 15 && wordCount != 18 && wordCount != 21 && wordCount != 24 {
		return nil, ErrInvalidMnemonicWordCount
	}

	// Convert words to indices
	indices := make([]int, wordCount)
	for i, word := range words {
		idx := slices.Index(wordlist, word)
		if idx < 0 {
			return nil, ErrInvalidMnemonic
		}
		indices[i] = idx
	}

	// Extract bits from indices
	totalBits := wordCount * 11
	checksumBits := wordCount / 3
	entropyBits := totalBits - checksumBits

	bits := make([]bool, totalBits)
	for i, idx := range indices {
		for j := 0; j < 11; j++ {
			bits[i*11+j] = (idx & (1 << (10 - j))) != 0
		}
	}

	// Extract entropy
	entropy := make([]byte, entropyBits/8)
	for i := 0; i < entropyBits; i++ {
		if bits[i] {
			entropy[i/8] |= 1 << (7 - (i % 8))
		}
	}

	// Verify checksum
	hash := sha256.Sum256(entropy)
	for i := 0; i < checksumBits; i++ {
		expected := (hash[0] & (1 << (7 - i))) != 0
		if bits[entropyBits+i] != expected {
			return nil, ErrInvalidMnemonic
		}
	}

	return entropy, nil
}

// MnemonicToSeed converts mnemonic words to a 64-byte seed using PBKDF2.
// Validates the mnemonic first. Passphrase is optional (use "" for none).
func MnemonicToSeed(words []string, passphrase string) (Seed, error) {
	if _, err := MnemonicToEntropy(words); err != nil {
		return nil, err
	}
	mnemonic := strings.Join(words, " ")
	salt := "mnemonic" + passphrase
	return Seed(pbkdf2.Key([]byte(mnemonic), []byte(salt), 2048, SeedLengthBytes, sha512.New)), nil
}

func NewEntropy(bits int) (Entropy, error) {
	if bits%EntropyStepBits != 0 || bits < MinEntropyBits || bits > MaxEntropyBits {
		return nil, fmt.Errorf("invalid entropy size: %d", bits)
	}

	bytes := bits / 8
	entropy := make(Entropy, bytes)

	_, err := rand.Read(entropy)
	if err != nil {
		return nil, err
	}

	return entropy, nil
}
