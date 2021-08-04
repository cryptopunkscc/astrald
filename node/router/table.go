package router

import (
	"encoding/json"
	"errors"
	_fs "github.com/cryptopunkscc/astrald/node/fs"
	"github.com/cryptopunkscc/astrald/node/net"
	"log"
	"os"
)

// Table maps identities to their addresses in various networks
type Table struct {
	entries map[string][]net.Addr
	fs      *_fs.Filesystem
}

type cacheAddr struct {
	Network string
	Address string
}

type cache map[string][]cacheAddr

// NewTable instantiates a new routing table
func NewTable(fs *_fs.Filesystem) *Table {
	t := &Table{
		fs:      fs,
		entries: make(map[string][]net.Addr),
	}
	err := t.load()
	if err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			log.Println("error reading table cache:", err)
		}
	}
	return t
}

// Add adds an endpoint to an identity, ignoring duplicates.
func (table *Table) Add(nodeID string, endpoint net.Addr) {
	// Make sure the map is initialized
	if table.entries == nil {
		table.entries = make(map[string][]net.Addr)
	}
	if _, found := table.entries[nodeID]; !found {
		table.entries[nodeID] = make([]net.Addr, 0)
	}

	// Search for duplicates
	for _, e := range table.entries[nodeID] {
		if e.String() == endpoint.String() {
			return
		}
	}

	// Add endpoint to the list
	table.entries[nodeID] = append(table.entries[nodeID], endpoint)

	// Persist the table
	table.save()
}

// Find fetches a list of known endpoints for the provided Identity
func (table *Table) Find(nodeID string) []net.Addr {
	if table.entries == nil {
		return nil
	}
	return table.entries[nodeID]
}

func (table *Table) load() error {
	bytes, err := table.fs.Read("table")
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
			table.Add(id, net.MakeAddr(a.Network, a.Address))
		}
	}

	return nil
}

func (table *Table) save() error {
	c := make(cache)
	for id, addrs := range table.entries {
		c[id] = make([]cacheAddr, len(addrs))
		for i := range addrs {
			c[id][i] = cacheAddr{
				Network: addrs[i].Network(),
				Address: addrs[i].String(),
			}
		}
	}
	bytes, _ := json.Marshal(&c)

	return table.fs.Write("table", bytes)
}
