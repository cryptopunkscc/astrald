package tcp

import (
	"context"
	"github.com/cryptopunkscc/astrald/node/net"
	"log"
	go_net "net"
	"time"
)

const tcpPort = 1791

func (drv *driver) Listen(ctx context.Context) (<-chan net.Conn, error) {
	output := make(chan net.Conn)

	go func() {
		defer close(output)

		ifaces := make(map[string]*listenerGroup)

		for {
			updated, err := getInterfaces()
			if err != nil {
				panic(err)
			}

			// Check which interfaces went down
			for name, iface := range ifaces {
				if _, found := updated[name]; found {
					continue
				}

				iface.close()
				log.Println("down", name)
				delete(ifaces, name)
			}

			// Check which interfaces went up
			for name, _ := range updated {
				if _, found := ifaces[name]; found {
					continue
				}

				log.Println("up", name)

				ifaces[name] = newListenerGroup()
				go func(name string) {
					for conn := range ifaces[name].Conns() {
						output <- conn
					}
				}(name)
			}

			// Check all current addresses
			for ifaceName, iface := range updated {
				addrs, err := iface.Addrs()
				if err != nil {
					continue
				}

				for _, addr := range addrs {
					ifaces[ifaceName].add(addr)
				}
			}

			select {
			case <-ctx.Done():
				return
			case <-time.After(time.Second):
			}
		}
	}()

	return output, nil
}

func getInterfaces() (map[string]go_net.Interface, error) {
	ifaces, err := go_net.Interfaces()
	if err != nil {
		return nil, err
	}

	m := make(map[string]go_net.Interface)
	for _, iface := range ifaces {
		m[iface.Name] = iface
	}
	return m, err
}
