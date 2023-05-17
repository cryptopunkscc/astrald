package tor

import (
	"context"
)

type Key []byte

const privateKeyFileName = "tor.key"

func (drv *Driver) getPrivateKey() (Key, error) {
	if key, err := drv.loadPrivateKey(); err == nil {
		return key, err
	}

	key, err := drv.generatePrivateKey()
	if err != nil {
		return nil, err
	}

	err = drv.savePrivateKey(key)

	return key, err
}

func (drv *Driver) loadPrivateKey() (Key, error) {
	return drv.assets.Read(privateKeyFileName)
}

func (drv *Driver) savePrivateKey(key Key) error {
	return drv.assets.Write(privateKeyFileName, key)
}

func (drv *Driver) generatePrivateKey() (Key, error) {
	service, err := drv.listen(context.Background(), nil)
	if err != nil {
		return nil, err
	}

	defer service.Close()

	return service.PrivateKey(), err
}
