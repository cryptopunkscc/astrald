package tor

import (
	"github.com/cryptopunkscc/astrald/infra/tor/tc"
)

const privateKeyFileName = "tor.key"

func (tor *Tor) getPrivateKey() (string, error) {
	if key, err := tor.loadPrivateKey(); err == nil {
		return key, err
	}

	key, err := tor.generatePrivateKey()
	if err != nil {
		return "", err
	}

	err = tor.savePrivateKey(key)

	return key, err
}

func (tor *Tor) loadPrivateKey() (string, error) {
	bytes, err := tor.store.LoadBytes(privateKeyFileName)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

func (tor *Tor) savePrivateKey(key string) error {
	return tor.store.StoreBytes(privateKeyFileName, []byte(key))
}

func (tor *Tor) generatePrivateKey() (string, error) {
	ctl, err := tor.connect()
	if err != nil {
		return "", err
	}
	defer ctl.Close()

	onion, err := ctl.AddOnion(tc.KeyNewBest, tc.Port(1024, "localhost:1024"))
	if err != nil {
		return "", err
	}

	return onion.PrivateKey, err
}
