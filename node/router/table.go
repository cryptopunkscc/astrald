package router

import (
	"github.com/cryptopunkscc/astrald/node/net"
)

// Table maps identities to their addresses in various networks
type Table struct {
	entries map[string][]net.Endpoint
}

// NewTable instantiates a new routing table
func NewTable() *Table {
	return &Table{entries: make(map[string][]net.Endpoint)}
}

// Add adds an endpoint to an identity, ignoring duplicates.
func (table *Table) Add(nodeID string, endpoint net.Endpoint) {
	// Make sure the map is initialized
	if table.entries == nil {
		table.entries = make(map[string][]net.Endpoint)
	}
	if _, found := table.entries[nodeID]; !found {
		table.entries[nodeID] = make([]net.Endpoint, 0)
	}

	// Search for duplicates
	for _, e := range table.entries[nodeID] {
		if e == endpoint {
			return
		}
	}

	// Add endpoint to the list
	table.entries[nodeID] = append(table.entries[nodeID], endpoint)
}

// Find fetches a list of known endpoints for the provided Identity
func (table *Table) Find(nodeID string) []net.Endpoint {
	if table.entries == nil {
		return nil
	}
	return table.entries[nodeID]
}
