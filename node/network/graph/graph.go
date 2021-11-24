package graph

import (
	"bytes"
	"errors"
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/infra"
	"io"
	"sync"
)

var _ Resolver = &Graph{}

type Graph struct {
	mu   sync.Mutex
	info map[string]*Info
}

func New() *Graph {
	return &Graph{
		info: make(map[string]*Info),
	}
}

// Resolve returns all known addresses of a node via channel
func (graph *Graph) Resolve(nodeID id.Identity) <-chan infra.Addr {
	graph.mu.Lock()
	defer graph.mu.Unlock()

	info := graph.info[nodeID.PublicKeyHex()]

	// convert to channel
	ch := make(chan infra.Addr, len(info.Addresses))
	for _, addr := range info.Addresses {
		ch <- addr
	}
	close(ch)

	return ch
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

func (graph *Graph) AddAddr(nodeID id.Identity, addr infra.Addr) {
	graph.mu.Lock()
	defer graph.mu.Unlock()

	graph.addAddr(nodeID, addr)
}

func (graph *Graph) AddInfo(info *Info) {
	graph.mu.Lock()
	defer graph.mu.Unlock()

	for _, addr := range info.Addresses {
		graph.addAddr(info.Identity, addr)
	}
}

func (graph *Graph) AddPacked(packed []byte) error {
	buf := bytes.NewReader(packed)
	for {
		r, err := Read(buf)
		if err != nil {
			if errors.Is(err, io.EOF) {
				return nil
			}
			return err
		}
		graph.AddInfo(r)
	}
}

func (graph *Graph) Pack() []byte {
	graph.mu.Lock()
	defer graph.mu.Unlock()

	buf := &bytes.Buffer{}
	for _, r := range graph.info {
		_ = Write(buf, r)
	}
	return buf.Bytes()
}

func (graph *Graph) addAddr(nodeID id.Identity, addr infra.Addr) {
	hex := nodeID.PublicKeyHex()

	if _, found := graph.info[hex]; !found {
		graph.info[hex] = NewInfo(nodeID)
	}

	graph.info[hex].Add(addr)
}
