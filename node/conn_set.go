package node

import (
	"errors"
	"sync"
)

type ConnSet struct {
	mu    sync.Mutex
	items []*MonitoredConn
	incl  map[*MonitoredConn]struct{}
}

func NewConnSet() *ConnSet {
	return &ConnSet{
		items: make([]*MonitoredConn, 0),
		incl:  make(map[*MonitoredConn]struct{}),
	}
}

func (set *ConnSet) Add(conn *MonitoredConn) error {
	set.mu.Lock()
	defer set.mu.Unlock()

	if _, found := set.incl[conn]; found {
		return errors.New("already added")
	}

	set.items = append(set.items, conn)
	set.incl[conn] = struct{}{}
	return nil
}

func (set *ConnSet) Remove(conn *MonitoredConn) error {
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

func (set *ConnSet) All() []*MonitoredConn {
	set.mu.Lock()
	defer set.mu.Unlock()

	var clone = make([]*MonitoredConn, len(set.items))
	copy(clone, set.items)
	return clone
}

func (set *ConnSet) Count() int {
	return len(set.items)
}

func (set *ConnSet) Find(id int) *MonitoredConn {
	set.mu.Lock()
	defer set.mu.Unlock()

	for _, conn := range set.items {
		if conn.ID() == id {
			return conn
		}
	}

	return nil
}
