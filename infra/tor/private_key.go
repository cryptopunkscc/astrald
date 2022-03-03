package tor

import (
	"context"
)

const privateKeyFileName = "tor.key"

func (tor *Tor) getPrivateKey() (Key, error) {
	if key, err := tor.loadPrivateKey(); err == nil {
		return key, err
	}

	key, err := tor.generatePrivateKey()
	if err != nil {
		return nil, err
	}

	err = tor.savePrivateKey(key)

	return key, err
}

func (tor *Tor) loadPrivateKey() (Key, error) {
	return tor.store.LoadBytes(privateKeyFileName)
}

func (tor *Tor) savePrivateKey(key Key) error {
	return tor.store.StoreBytes(privateKeyFileName, key)
}

func (tor *Tor) generatePrivateKey() (Key, error) {
	service, err := tor.backend.Listen(context.Background(), nil)
	if err != nil {
		return nil, err
	}

	defer service.Close()

	return service.PrivateKey(), err
}
