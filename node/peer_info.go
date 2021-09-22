package node

import (
	"encoding/json"
	"errors"
	"github.com/cryptopunkscc/astrald/net"
	_fs "github.com/cryptopunkscc/astrald/node/fs"
	"log"
	"os"
	"sync"
)

const peerInfoFilename = "peers"

type AddrList []net.Addr

// PeerInfo maps identities to their addresses in various networks
type PeerInfo struct {
	entries map[string]AddrList
	fs      *_fs.Filesystem
	mu      sync.Mutex
}

type cacheAddr struct {
	Network string
	Address string
}

type cache map[string][]cacheAddr

// NewPeerInfo instantiates a new routing table
func NewPeerInfo(fs *_fs.Filesystem) *PeerInfo {
	t := &PeerInfo{
		fs:      fs,
		entries: make(map[string]AddrList),
	}
	err := t.load()
	if err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			log.Println("error reading table cache:", err)
		}
	}

	log.Println(len(t.entries), "peer(s) in the cache")

	return t
}

// Add adds an endpoint to an identity, ignoring duplicates.
func (info *PeerInfo) Add(nodeID string, endpoint net.Addr) {
	info.mu.Lock()
	defer info.mu.Unlock()

	// Make sure the map is initialized
	if info.entries == nil {
		info.entries = make(map[string]AddrList)
	}
	if _, found := info.entries[nodeID]; !found {
		info.entries[nodeID] = make(AddrList, 0)
	}

	// Search for duplicates
	for _, e := range info.entries[nodeID] {
		if e.String() == endpoint.String() {
			return
		}
	}

	// Add endpoint to the list
	info.entries[nodeID] = append(info.entries[nodeID], endpoint)

	// Persist the table
	err := info.save()
	if err != nil {
		log.Println("error saving peer info:", err)
	}
}

// NodeAddr fetches a list of known endpoints for the provided Identity
func (info *PeerInfo) NodeAddr(nodeID string) AddrList {
	info.mu.Lock()
	defer info.mu.Unlock()

	if info.entries == nil {
		return AddrList{}
	}
	return info.entries[nodeID]
}

func (info *PeerInfo) Each() <-chan *AddrList {
	info.mu.Lock()
	defer info.mu.Unlock()

	ch := make(chan *AddrList, len(info.entries))

	for _, e := range info.entries {
		ch <- &e
	}

	return ch
}

func (info *PeerInfo) load() error {
	bytes, err := info.fs.Read(peerInfoFilename)
	if err != nil {
		return err
	}

	c := make(cache)
	err = json.Unmarshal(bytes, &c)
	if err != nil {
		return err
	}

	for id, addrs := range c {
		for _, a := range addrs {
			info.Add(id, net.MakeAddr(a.Network, a.Address))
		}
	}

	return nil
}

func (info *PeerInfo) save() error {
	c := make(cache)
	for id, addrs := range info.entries {
		c[id] = make([]cacheAddr, len(addrs))
		for i := range addrs {
			c[id][i] = cacheAddr{
				Network: addrs[i].Network(),
				Address: addrs[i].String(),
			}
		}
	}
	bytes, _ := json.MarshalIndent(&c, "", "  ")

	return info.fs.Write(peerInfoFilename, bytes)
}

func (list AddrList) OnlyNet(netName string) AddrList {
	res := make(AddrList, 0)
	for _, addr := range list {
		if addr.Network() == netName {
			res = append(res, addr)
		}
	}
	return res
}
