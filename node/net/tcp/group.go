package tcp

import (
	"github.com/cryptopunkscc/astrald/node/net"
	"github.com/cryptopunkscc/astrald/node/net/ip"
	"log"
	_net "net"
	"strconv"
	"sync"
)

type listenerGroup struct {
	listeners map[string]_net.Listener
	conns     chan net.Conn
	mu        sync.Mutex
}

func newListenerGroup() *listenerGroup {
	return &listenerGroup{
		listeners: make(map[string]_net.Listener),
		conns:     make(chan net.Conn),
	}
}

func (grp *listenerGroup) Conns() <-chan net.Conn {
	return grp.conns
}

func (grp *listenerGroup) add(addr _net.Addr) {
	grp.mu.Lock()
	defer grp.mu.Unlock()

	if _, found := grp.listeners[addr.String()]; found {
		return
	}

	ip, _ := ip.SplitIPMask(addr.String())
	hostPort := _net.JoinHostPort(ip, strconv.Itoa(tcpPort))

	listener, err := _net.Listen("tcp", hostPort)
	if err != nil {
		return
	}

	grp.listeners[addr.String()] = listener

	go func(addr _net.Addr) {
		log.Println("listen tcp", hostPort)

		defer func() {
			grp.mu.Lock()
			defer grp.mu.Unlock()
			if grp.listeners != nil {
				grp.listeners[addr.String()] = nil
			}
		}()

		for {
			conn, err := listener.Accept()
			if err != nil {
				break
			}
			grp.conns <- net.WrapConn(conn, false)
		}

		log.Println("closed tcp", hostPort)
	}(addr)
}

func (grp *listenerGroup) close() {
	grp.mu.Lock()
	defer grp.mu.Unlock()

	for _, l := range grp.listeners {
		l.Close()
	}
	close(grp.conns)
	grp.listeners = nil
}
