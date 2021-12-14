package tor

import (
	"github.com/cryptopunkscc/astrald/infra/tor/tc"
	"io/ioutil"
	"os"
	"path"
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
	bytes, err := ioutil.ReadFile(tor.privateKeyPath())
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

func (tor *Tor) savePrivateKey(key string) error {
	return ioutil.WriteFile(tor.privateKeyPath(), []byte(key), 0600)
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

func (tor *Tor) privateKeyPath() string {
	if tor.config.DataDir != "" {
		return path.Join(tor.config.DataDir, privateKeyFileName)
	}
	home, err := os.UserHomeDir()
	if err == nil {
		return path.Join(home, privateKeyFileName)
	}
	return privateKeyFileName
}
