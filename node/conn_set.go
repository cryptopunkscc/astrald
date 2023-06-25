package node

import (
	"errors"
	"sync"
)

type ConnSet struct {
	mu    sync.Mutex
	items []*Conn
	incl  map[*Conn]struct{}
}

func NewConnSet() *ConnSet {
	return &ConnSet{
		items: make([]*Conn, 0),
		incl:  make(map[*Conn]struct{}),
	}
}

func (set *ConnSet) Add(conn *Conn) error {
	set.mu.Lock()
	defer set.mu.Unlock()

	if _, found := set.incl[conn]; found {
		return errors.New("already added")
	}

	set.items = append(set.items, conn)
	set.incl[conn] = struct{}{}
	return nil
}

func (set *ConnSet) Remove(conn *Conn) error {
	set.mu.Lock()
	defer set.mu.Unlock()

	if _, found := set.incl[conn]; !found {
		return errors.New("not found")
	}

	delete(set.incl, conn)
	for i := range set.items {
		if set.items[i] == conn {
			set.items = append(set.items[:i], set.items[i+1:]...)
			return nil
		}
	}
	return nil
}

func (set *ConnSet) All() []*Conn {
	set.mu.Lock()
	defer set.mu.Unlock()

	var clone = make([]*Conn, len(set.items))
	copy(clone, set.items)
	return clone
}

func (set *ConnSet) Count() int {
	return len(set.items)
}

func (set *ConnSet) FindByRemotePort(remotePort int) *Conn {
	set.mu.Lock()
	defer set.mu.Unlock()

	for _, conn := range set.items {
		if conn.RemotePort() == remotePort {
			return conn
		}
	}

	return nil
}

func (set *ConnSet) FindByLocalPort(localPort int) *Conn {
	set.mu.Lock()
	defer set.mu.Unlock()

	for _, conn := range set.items {
		if conn.LocalPort() == localPort {
			return conn
		}
	}

	return nil
}
