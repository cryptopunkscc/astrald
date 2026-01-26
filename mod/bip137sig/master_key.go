package bip137sig

import (
	"github.com/btcsuite/btcd/btcutil/hdkeychain"
	"github.com/btcsuite/btcd/chaincfg"
)

func MasterKeyFromSeed(seed Seed) (*hdkeychain.ExtendedKey, error) {
	return hdkeychain.NewMaster(seed.Data, &chaincfg.MainNetParams)
}
