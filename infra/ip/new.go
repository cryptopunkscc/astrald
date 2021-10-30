package ip

import (
	"context"
	"net"
	"time"
)

type Interface struct {
	name string
	gone chan struct{}
}

type Addr struct {
	*net.IPNet
	gone chan struct{}
}

func (addr *Addr) IsPrivate() bool {
	return addr.IPNet.IP.IsPrivate()
}

func (addr *Addr) String() string {
	return addr.IPNet.String()
}

func (addr *Addr) Gone() <-chan struct{} {
	return addr.gone
}

func NewAddress(addr *net.IPNet) *Addr {
	return &Addr{
		IPNet: addr,
		gone:  make(chan struct{}),
	}
}

func (addr *Addr) close() {
	close(addr.gone)
}

func NewInterface(name string) *Interface {
	return &Interface{
		name: name,
		gone: make(chan struct{}),
	}
}

func (iface *Interface) Name() string {
	return iface.name
}

func (iface *Interface) Gone() <-chan struct{} {
	return iface.gone
}

func (iface *Interface) addressMap() (map[string]*net.IPNet, error) {
	var err error

	netIface, err := net.InterfaceByName(iface.name)
	if err != nil {
		return nil, err
	}

	addrs, err := netIface.Addrs()
	if err != nil {
		return nil, err
	}

	m := make(map[string]*net.IPNet)
	for _, addr := range addrs {
		ipAddr, ok := addr.(*net.IPNet)
		if !ok {
			continue
		}

		m[ipAddr.String()] = ipAddr
	}
	return m, nil
}

func (iface *Interface) WatchAddrs(ctx context.Context) <-chan *Addr {
	outCh := make(chan *Addr)

	go func() {
		defer close(outCh)

		oldAddrs := make(map[string]*Addr)

		for {
			newAddrs, err := iface.addressMap()
			if err != nil {
				return
			}

			// check what's gone
			for name, addr := range oldAddrs {
				if _, found := newAddrs[name]; !found {
					addr.close()
					delete(oldAddrs, name)
				}
			}

			// check what's new
			for name, addr := range newAddrs {
				if _, found := oldAddrs[name]; !found {
					oldAddrs[name] = NewAddress(addr)
					outCh <- oldAddrs[name]
				}
			}

			select {
			case <-ctx.Done():
				return
			case <-time.After(pollInterval):
			}
		}
	}()

	return outCh
}

func (iface *Interface) close() {
	close(iface.gone)
}

func WatchInterfaces(ctx context.Context) <-chan *Interface {
	outCh := make(chan *Interface)

	var handlers = make(map[string]*Interface)

	go func() {
		defer close(outCh)

		for {
			interfaces, err := interfaceMap()
			if err != nil {
				return
			}

			// First check what's gone
			for name, handler := range handlers {
				if _, found := interfaces[name]; !found {
					handler.close()
					delete(handlers, name)
				}
			}

			// Then check what's new
			for name, _ := range interfaces {
				if _, found := handlers[name]; !found {
					handlers[name] = NewInterface(name)
					outCh <- handlers[name]
				}
			}

			select {
			case <-ctx.Done():
				return
			case <-time.After(pollInterval):
			}
		}
	}()

	return outCh
}

func interfaceMap() (map[string]net.Interface, error) {
	interfaces, err := net.Interfaces()
	if err != nil {
		return nil, err
	}

	m := make(map[string]net.Interface)
	for _, iface := range interfaces {
		m[iface.Name] = iface
	}
	return m, nil
}
