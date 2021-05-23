package router

import (
	"github.com/cryptopunkscc/astrald/node/net"
	"log"
)

// Table maps identities to their addresses in various networks
type Table struct {
	entries map[string][]net.Addr
}

// NewTable instantiates a new routing table
func NewTable() *Table {
	return &Table{entries: make(map[string][]net.Addr)}
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

	log.Println("new endpoint", endpoint.Network(), endpoint.String(), "for", nodeID)

	// Add endpoint to the list
	table.entries[nodeID] = append(table.entries[nodeID], endpoint)
}

// Find fetches a list of known endpoints for the provided Identity
func (table *Table) Find(nodeID string) []net.Addr {
	if table.entries == nil {
		return nil
	}
	return table.entries[nodeID]
}
