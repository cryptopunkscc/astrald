package contacts

import (
	"bytes"
	"errors"
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/infra"
	"github.com/cryptopunkscc/astrald/node/storage"
	"gopkg.in/yaml.v2"
	"io"
	"log"
	"sync"
)

const graphKey = "graph"
const aliasesKey = "aliases"

var _ Resolver = &Manager{}

type Manager struct {
	store   storage.Store
	mu      sync.Mutex
	info    map[string]*Contact
	aliases map[string]string
}

func New(store storage.Store) *Manager {
	c := &Manager{
		store:   store,
		info:    make(map[string]*Contact),
		aliases: make(map[string]string),
	}

	c.load()

	return c
}

// Resolve returns all known addresses of a node via channel
func (c *Manager) Resolve(nodeID id.Identity) <-chan infra.Addr {
	c.mu.Lock()
	defer c.mu.Unlock()

	info := c.info[nodeID.PublicKeyHex()]
	if info == nil {
		ch := make(chan infra.Addr)
		close(ch)
		return ch
	}

	// convert to channel
	ch := make(chan infra.Addr, len(info.Addresses))
	for _, addr := range info.Addresses {
		ch <- addr
	}
	close(ch)

	return ch
}

func (c *Manager) ResolveAlias(alias string) (string, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	s, ok := c.aliases[alias]
	return s, ok
}

func (c *Manager) ResolveIdentity(str string) (id.Identity, error) {
	if id, err := id.ParsePublicKeyHex(str); err == nil {
		return id, nil
	}

	target, found := c.ResolveAlias(str)

	if !found {
		return id.Identity{}, errors.New("unknown identity")
	}
	if str == target {
		return id.Identity{}, errors.New("circular alias")
	}
	return c.ResolveIdentity(target)
}

func (c *Manager) GetAlias(identity id.Identity) string {
	if info, ok := c.info[identity.PublicKeyHex()]; ok {
		return info.Alias
	}
	return ""
}

// Identities returns a closed channel populated with all known node IDs
func (c *Manager) Identities() <-chan id.Identity {
	c.mu.Lock()
	defer c.mu.Unlock()

	ch := make(chan id.Identity, len(c.info))
	for _, info := range c.info {
		ch <- info.Identity
	}
	close(ch)

	return ch
}

func (c *Manager) AddInfo(info *Contact) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.addInfo(info)
	c.save()
}

func (c *Manager) AddAddr(nodeID id.Identity, addr infra.Addr) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.addAddr(nodeID, addr)
	c.save()
}

func (c *Manager) addInfo(info *Contact) {
	for _, addr := range info.Addresses {
		c.addAddr(info.Identity, addr)
	}

	if info.Alias != "" {
		hex := info.Identity.PublicKeyHex()
		c.info[hex].Alias = info.Alias
		c.aliases[info.Alias] = info.Identity.PublicKeyHex()
	}
}

func (c *Manager) addAddr(nodeID id.Identity, addr infra.Addr) {
	hex := nodeID.PublicKeyHex()

	if _, found := c.info[hex]; !found {
		c.info[hex] = NewContact(nodeID)
	}

	c.info[hex].Add(addr)
}

func (c *Manager) load() error {
	if err := c.loadGraph(); err != nil {
		log.Println("error loading graph:", err)
	}
	c.loadAliases()
	return nil
}

func (c *Manager) loadGraph() error {
	packed, err := c.store.LoadBytes(graphKey)
	if err != nil {
		return err
	}
	buf := bytes.NewReader(packed)
	for {
		info, err := readInfo(buf)
		if err != nil {
			if errors.Is(err, io.EOF) {
				return nil
			}
			return err
		}
		c.addInfo(info)
	}
}

func (c *Manager) loadAliases() error {
	bytes, err := c.store.LoadBytes(aliasesKey)
	if err != nil {
		return err
	}

	return yaml.Unmarshal(bytes, &c.aliases)
}

func (c *Manager) pack() []byte {
	buf := &bytes.Buffer{}
	for _, r := range c.info {
		_ = writeInfo(buf, r)
	}
	return buf.Bytes()
}

func (c *Manager) save() error {
	if err := c.saveGraph(); err != nil {
		log.Println("error saving graph:", err)
	}
	c.saveAliases()
	return nil
}

func (c *Manager) saveGraph() error {
	return c.store.StoreBytes(graphKey, c.pack())
}

func (c *Manager) saveAliases() {
	bytes, err := yaml.Marshal(c.aliases)
	if err != nil {
		return
	}

	c.store.StoreBytes(aliasesKey, bytes)
}
