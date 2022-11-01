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
	mu   sync.Mutex
	subs map[Listener]Unsubscribe
}

func NewSubscriptions() *Subscriptions {
	return &Subscriptions{subs: map[Listener]Unsubscribe{}}
}

func (s *Subscriptions) Notify(data interface{}) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for subscriber := range s.subs {
		subscriber <- data
	}
}

func (s *Subscriptions) Subscribe(c Listener) (unsub Unsubscribe) {
	unsub = func() {
		s.mu.Lock()
		delete(s.subs, c)
		s.mu.Unlock()
	}
	s.mu.Lock()
	s.subs[c] = unsub
	s.mu.Unlock()
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
