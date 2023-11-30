package router

import "github.com/cryptopunkscc/astrald/net"

// LocalRouter is a router that routes queries for a single local identity
type LocalRouter interface {
	net.Router
	AddRoute(name string, target net.Router) error
	RemoveRoute(name string) error
	Routes() []LocalRoute
	Match(query string) net.Router
}

type LocalRoute struct {
	Name   string
	Target net.Router
}
