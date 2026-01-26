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

// reference https://github.com/bitcoin/bips/blob/master/bip-0039.mediawiki

// EntropyToMnemonic converts entropy bytes to mnemonic words.
// Entropy must be 16, 20, 24, 28, or 32 bytes (128-256 bits in 32-bit increments).
func EntropyToMnemonic(entropy []byte) ([]string, error) {
	entropyBits := len(entropy) * 8
	if entropyBits < 128 || entropyBits > 256 || entropyBits%32 != 0 {
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
func MnemonicToEntropy(words []string) ([]byte, error) {
	wordCount := len(words)
	if wordCount != 12 && wordCount != 15 && wordCount != 18 && wordCount != 21 && wordCount != 24 {
		return nil, errors.New("mnemonic must be 12, 15, 18, 21, or 24 words")
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
// Passphrase is optional (use "" for none).
func MnemonicToSeed(words []string, passphrase string) []byte {
	mnemonic := strings.Join(words, " ")
	salt := "mnemonic" + passphrase
	return pbkdf2.Key([]byte(mnemonic), []byte(salt), 2048, 64, sha512.New)
}

func NewEntropy(bits int) ([]byte, error) {
	if bits%32 != 0 || bits < 128 || bits > 256 {
		return nil, fmt.Errorf("invalid entropy size: %d", bits)
	}

	bytes := bits / 8
	entropy := make([]byte, bytes)

	_, err := rand.Read(entropy)
	if err != nil {
		return nil, err
	}

	return entropy, nil
}
