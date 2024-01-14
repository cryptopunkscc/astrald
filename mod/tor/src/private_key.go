package tor

import (
	"context"
)

type Key []byte

const privateKeyFileName = "tor.key"

func (mod *Module) getPrivateKey() (Key, error) {
	if key, err := mod.loadPrivateKey(); err == nil {
		return key, err
	}

	key, err := mod.generatePrivateKey()
	if err != nil {
		return nil, err
	}

	err = mod.savePrivateKey(key)

	return key, err
}

func (mod *Module) loadPrivateKey() (Key, error) {
	return mod.assets.Read(privateKeyFileName)
}

func (mod *Module) savePrivateKey(key Key) error {
	return mod.assets.Write(privateKeyFileName, key)
}

func (mod *Module) generatePrivateKey() (Key, error) {
	service, err := mod.server.listen(context.Background(), nil)
	if err != nil {
		return nil, err
	}

	defer service.Close()

	return service.PrivateKey(), err
}
