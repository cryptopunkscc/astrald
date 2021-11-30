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

var _ Resolver = &Contacts{}

type Contacts struct {
	store   storage.Store
	mu      sync.Mutex
	info    map[string]*Info
	aliases map[string]string
}

func New(store storage.Store) *Contacts {
	c := &Contacts{
		store:   store,
		info:    make(map[string]*Info),
		aliases: make(map[string]string),
	}

	c.load()

	return c
}

// Resolve returns all known addresses of a node via channel
func (c *Contacts) Resolve(nodeID id.Identity) <-chan infra.Addr {
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

func (c *Contacts) ResolveAlias(alias string) (string, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	s, ok := c.aliases[alias]
	return s, ok
}

func (c *Contacts) GetAlias(identity id.Identity) string {
	if info, ok := c.info[identity.PublicKeyHex()]; ok {
		return info.Alias
	}
	return ""
}

// Identities returns a closed channel populated with all known node IDs
func (c *Contacts) Identities() <-chan id.Identity {
	c.mu.Lock()
	defer c.mu.Unlock()

	ch := make(chan id.Identity, len(c.info))
	for _, info := range c.info {
		ch <- info.Identity
	}
	close(ch)

	return ch
}

func (c *Contacts) AddInfo(info *Info) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.addInfo(info)
	c.save()
}

func (c *Contacts) AddAddr(nodeID id.Identity, addr infra.Addr) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.addAddr(nodeID, addr)
	c.save()
}

func (c *Contacts) addInfo(info *Info) {
	for _, addr := range info.Addresses {
		c.addAddr(info.Identity, addr)
	}

	if info.Alias != "" {
		hex := info.Identity.PublicKeyHex()
		c.info[hex].Alias = info.Alias
		c.aliases[info.Alias] = info.Identity.PublicKeyHex()
	}
}

func (c *Contacts) addAddr(nodeID id.Identity, addr infra.Addr) {
	hex := nodeID.PublicKeyHex()

	if _, found := c.info[hex]; !found {
		c.info[hex] = NewInfo(nodeID)
	}

	c.info[hex].Add(addr)
}

func (c *Contacts) load() error {
	if err := c.loadGraph(); err != nil {
		log.Println("error loading graph:", err)
	}
	c.loadAliases()
	return nil
}

func (c *Contacts) loadGraph() error {
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

func (c *Contacts) loadAliases() error {
	bytes, err := c.store.LoadBytes(aliasesKey)
	if err != nil {
		return err
	}

	return yaml.Unmarshal(bytes, &c.aliases)
}

func (c *Contacts) pack() []byte {
	buf := &bytes.Buffer{}
	for _, r := range c.info {
		_ = writeInfo(buf, r)
	}
	return buf.Bytes()
}

func (c *Contacts) save() error {
	if err := c.saveGraph(); err != nil {
		log.Println("error saving graph:", err)
	}
	c.saveAliases()
	return nil
}

func (c *Contacts) saveGraph() error {
	return c.store.StoreBytes(graphKey, c.pack())
}

func (c *Contacts) saveAliases() {
	bytes, err := yaml.Marshal(c.aliases)
	if err != nil {
		return
	}

	c.store.StoreBytes(aliasesKey, bytes)
}
