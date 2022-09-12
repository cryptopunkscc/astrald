package warpdrive

import (
	"context"
	"encoding/json"
	"io"
	"log"
	"sync"
)

type Unsubscribe func()
type Listener chan<- interface{}

type Subscriptions struct {
	sync.Mutex
	Set map[Listener]Unsubscribe
}

func NewSubscriptions() *Subscriptions {
	return &Subscriptions{Set: map[Listener]Unsubscribe{}}
}

func (s *Subscriptions) Subscribe(c Listener) (unsub Unsubscribe) {
	unsub = func() {
		s.Lock()
		delete(s.Set, c)
		s.Unlock()
	}
	s.Lock()
	s.Set[c] = unsub
	s.Unlock()
	return
}

func NewListener(ctx context.Context, w io.WriteCloser) (listener Listener) {
	c := make(chan interface{}, 1024)
	listener = c
	e := json.NewEncoder(w)
	go func() {
		for {
			select {
			case <-ctx.Done():
				_ = w.Close()
				return
			case i, ok := <-c:
				if !ok {
					return
				}
				var err error
				switch v := i.(type) {
				case []byte:
					v = append(v, '\n')
					_, err = w.Write(v)
				default:
					err = e.Encode(i)
				}
				if err != nil {
					log.Println("Cannot write", err)
					return
				}
			}
		}
	}()
	return
}
