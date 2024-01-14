package apphost

import (
	"context"
	"github.com/cryptopunkscc/astrald/mod/apphost/proto"
	"net"
	"strings"
	"sync"
)

func (mod *Module) listen(ctx context.Context) <-chan net.Conn {
	var ch = make(chan net.Conn)
	var wg sync.WaitGroup
	mod.listeners = make([]net.Listener, 0)

	for _, endpoint := range mod.config.Listen {
		listener, err := proto.Listen(endpoint)

		if err != nil {
			mod.log.Error("listener %s error: %s", endpoint, err)
			continue
		}

		mod.listeners = append(mod.listeners, listener)

		mod.log.Infov(1, "listening on: %s %s", listener.Addr().Network(), listener.Addr().String())

		wg.Add(1)
		go func() {
			defer wg.Done()

			for {
				conn, err := listener.Accept()
				if err != nil {
					break
				}
				ch <- conn
			}
		}()
	}

	go func() {
		<-ctx.Done()
		for _, l := range mod.listeners {
			l.Close()
		}
	}()

	go func() {
		wg.Wait()
		close(ch)
	}()

	return ch
}

func (mod *Module) getListeners() string {
	var list = make([]string, 0)

	for _, l := range mod.listeners {
		list = append(list, l.Addr().Network()+":"+l.Addr().String())
	}

	return strings.Join(list, ";")
}
