package node

import (
	"errors"
	"fmt"
	"github.com/cryptopunkscc/astrald/auth/id"
	"os"
)

const defaultIdentityKey = "id"

func (node *CoreNode) setupIdentity() (err error) {
	//TODO: remove this function after migrations are done
	node.importIdentity(defaultIdentityKey)

	if len(node.config.Identity) > 0 {
		return node.loadConfigIdentity()
	}

	return node.loadDefaultIdentity()
}

func (node *CoreNode) loadConfigIdentity() error {
	identity, err := func() (i id.Identity, e error) {
		i, e = id.ParsePublicKeyHex(node.config.Identity)
		if e == nil {
			return
		}

		i, e = node.Tracker().IdentityByAlias(node.config.Identity)
		if e == nil {
			return
		}

		return id.Identity{}, errors.New("unknown identity")
	}()
	if err != nil {
		return err
	}

	if node.identity, err = node.keys.Find(identity); err != nil {
		return err
	}

	return nil
}

func (node *CoreNode) loadDefaultIdentity() error {
	count, err := node.keys.Count()
	if err != nil {
		return err
	}

	// if there are no keys in the store, this is the first run
	if count == 0 {
		var alias = "localnode"
		if hostname, err := os.Hostname(); err == nil {
			alias = hostname
		}

		node.identity, err = node.generateIdentity(alias)
	} else {
		node.identity, err = node.keys.First()
	}

	return err
}

func (node *CoreNode) generateIdentity(alias string) (id.Identity, error) {
	identity, err := id.GenerateIdentity()
	if err != nil {
		return id.Identity{}, err
	}

	node.tracker.SetAlias(identity, alias)

	return identity, node.keys.Save(identity)
}

// deprecated
func (node *CoreNode) importIdentity(name string) error {
	bytes, err := node.assets.Read(name)
	if err != nil {
		return err
	}

	identity, err := id.ParsePrivateKey(bytes)
	if err != nil {
		return err
	}

	fmt.Println("READ", identity.PublicKeyHex())

	return node.keys.Save(identity)
}
