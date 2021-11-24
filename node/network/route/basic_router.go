package route

import (
	"bytes"
	"errors"
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/infra"
	"io"
	"sync"
)

var _ Router = &BasicRouter{}

type BasicRouter struct {
	mu     sync.Mutex
	routes map[string]*Route
}

func NewBasicRouter() *BasicRouter {
	return &BasicRouter{
		routes: make(map[string]*Route),
	}
}

func (router *BasicRouter) Route(nodeID id.Identity) *Route {
	router.mu.Lock()
	defer router.mu.Unlock()

	return router.routes[nodeID.PublicKeyHex()]
}

func (router *BasicRouter) AddAddr(nodeID id.Identity, addr infra.Addr) {
	router.mu.Lock()
	defer router.mu.Unlock()

	router.addAddr(nodeID, addr)
}

func (router *BasicRouter) AddRoute(route *Route) {
	router.mu.Lock()
	defer router.mu.Unlock()

	for _, addr := range route.Addresses {
		router.addAddr(route.Identity, addr)
	}
}

func (router *BasicRouter) AddPacked(packed []byte) error {
	buf := bytes.NewReader(packed)
	for {
		r, err := Read(buf)
		if err != nil {
			if errors.Is(err, io.EOF) {
				return nil
			}
			return err
		}
		router.AddRoute(r)
	}
}

func (router *BasicRouter) Routes() <-chan *Route {
	router.mu.Lock()
	defer router.mu.Unlock()

	ch := make(chan *Route, len(router.routes))
	for _, r := range router.routes {
		ch <- r
	}
	close(ch)
	return ch
}

func (router *BasicRouter) RouteCount() int {
	return len(router.routes)
}

func (router *BasicRouter) Pack() []byte {
	router.mu.Lock()
	defer router.mu.Unlock()

	buf := &bytes.Buffer{}
	for _, r := range router.routes {
		_ = Write(buf, r)
	}
	return buf.Bytes()
}

func (router *BasicRouter) addAddr(nodeID id.Identity, addr infra.Addr) {
	hex := nodeID.PublicKeyHex()

	if _, found := router.routes[hex]; !found {
		router.routes[hex] = New(nodeID)
	}

	router.routes[hex].Add(addr)
}
