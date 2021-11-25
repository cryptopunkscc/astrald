package graph

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

var _ Resolver = &Graph{}

type Graph struct {
	store   storage.Store
	mu      sync.Mutex
	info    map[string]*Info
	aliases map[string]string
}

func New(store storage.Store) *Graph {
	graph := &Graph{
		store:   store,
		info:    make(map[string]*Info),
		aliases: make(map[string]string),
	}

	graph.load()

	return graph
}

// Resolve returns all known addresses of a node via channel
func (graph *Graph) Resolve(nodeID id.Identity) <-chan infra.Addr {
	graph.mu.Lock()
	defer graph.mu.Unlock()

	info := graph.info[nodeID.PublicKeyHex()]
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

func (graph *Graph) ResolveAlias(alias string) (string, bool) {
	graph.mu.Lock()
	defer graph.mu.Unlock()

	s, ok := graph.aliases[alias]
	return s, ok
}

func (graph *Graph) GetAlias(identity id.Identity) string {
	if info, ok := graph.info[identity.PublicKeyHex()]; ok {
		return info.Alias
	}
	return ""
}

// Nodes returns a closed channel populated with all known node IDs
func (graph *Graph) Nodes() <-chan id.Identity {
	graph.mu.Lock()
	defer graph.mu.Unlock()

	ch := make(chan id.Identity, len(graph.info))
	for _, info := range graph.info {
		ch <- info.Identity
	}
	close(ch)

	return ch
}

func (graph *Graph) AddInfo(info *Info) {
	graph.mu.Lock()
	defer graph.mu.Unlock()

	graph.addInfo(info)
	graph.save()
}

func (graph *Graph) AddAddr(nodeID id.Identity, addr infra.Addr) {
	graph.mu.Lock()
	defer graph.mu.Unlock()

	graph.addAddr(nodeID, addr)
	graph.save()
}

func (graph *Graph) addInfo(info *Info) {
	for _, addr := range info.Addresses {
		graph.addAddr(info.Identity, addr)
	}

	if info.Alias != "" {
		hex := info.Identity.PublicKeyHex()
		graph.info[hex].Alias = info.Alias
		graph.aliases[info.Alias] = info.Identity.PublicKeyHex()
	}
}

func (graph *Graph) addAddr(nodeID id.Identity, addr infra.Addr) {
	hex := nodeID.PublicKeyHex()

	if _, found := graph.info[hex]; !found {
		graph.info[hex] = NewInfo(nodeID)
	}

	graph.info[hex].Add(addr)
}

func (graph *Graph) load() error {
	if err := graph.loadGraph(); err != nil {
		log.Println("error loading graph:", err)
	}
	graph.loadAliases()
	return nil
}

func (graph *Graph) loadGraph() error {
	packed, err := graph.store.LoadBytes(graphKey)
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
		graph.addInfo(info)
	}
}

func (graph *Graph) loadAliases() error {
	bytes, err := graph.store.LoadBytes(aliasesKey)
	if err != nil {
		return err
	}

	return yaml.Unmarshal(bytes, &graph.aliases)
}

func (graph *Graph) pack() []byte {
	buf := &bytes.Buffer{}
	for _, r := range graph.info {
		_ = writeInfo(buf, r)
	}
	return buf.Bytes()
}

func (graph *Graph) save() error {
	if err := graph.saveGraph(); err != nil {
		log.Println("error saving graph:", err)
	}
	graph.saveAliases()
	return nil
}

func (graph *Graph) saveGraph() error {
	return graph.store.StoreBytes(graphKey, graph.pack())
}

func (graph *Graph) saveAliases() {
	bytes, err := yaml.Marshal(graph.aliases)
	if err != nil {
		return
	}

	graph.store.StoreBytes(aliasesKey, bytes)
}
