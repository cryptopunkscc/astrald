package node

import (
	"bytes"
	"errors"
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/node/route"
	"io"
	"sync"
)

var _ route.Router = &BasicRouter{}

type BasicRouter struct {
	mu     sync.Mutex
	routes map[string]*route.Route
}

func NewBasicRouter() *BasicRouter {
	return &BasicRouter{
		routes: make(map[string]*route.Route),
	}
}

func (router BasicRouter) Route(nodeID id.Identity) *route.Route {
	return router.routes[nodeID.PublicKeyHex()]
}

func (router *BasicRouter) AddRoute(r *route.Route) {
	router.mu.Lock()
	defer router.mu.Unlock()

	hex := r.Identity.PublicKeyHex()

	if _, found := router.routes[hex]; !found {
		router.routes[hex] = route.New(r.Identity)
	}

	_route := router.routes[hex]
	for _, a := range r.Addresses {
		_route.Add(a)
	}
}

func (router *BasicRouter) RouteCount() int {
	return len(router.routes)
}

func (router *BasicRouter) Pack() []byte {
	router.mu.Lock()
	defer router.mu.Unlock()

	buf := &bytes.Buffer{}
	for _, r := range router.routes {
		_ = route.Write(buf, r)
	}
	return buf.Bytes()
}

func (router *BasicRouter) AddPacked(packed []byte) error {
	buf := bytes.NewReader(packed)
	for {
		r, err := route.Read(buf)
		if err != nil {
			if errors.Is(err, io.EOF) {
				return nil
			}
			return err
		}
		router.AddRoute(r)
	}
}

func (router *BasicRouter) Routes() <-chan *route.Route {
	router.mu.Lock()
	defer router.mu.Unlock()

	ch := make(chan *route.Route, len(router.routes))
	for _, r := range router.routes {
		ch <- r
	}
	close(ch)
	return ch
}
