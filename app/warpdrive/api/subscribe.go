package api

import (
	"encoding/json"
	"io"
	"log"
	"sync"
)

type Unsubscribe func()

type Subscriptions struct {
	sync.Mutex
	Set map[chan<- interface{}]Unsubscribe
}

func NewSubscriptions() *Subscriptions {
	return &Subscriptions{Set: map[chan<- interface{}]Unsubscribe{}}
}

func (s *Subscriptions) SubscribeChan(c chan<- interface{}) (unsub Unsubscribe) {
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

func (s *Subscriptions) Subscribe(w io.WriteCloser) (unsub Unsubscribe) {
	c := WriteChannel(w)
	return s.SubscribeChan(c)
}
func WriteChannel(w io.WriteCloser) (writeChannel chan<- interface{}) {
	c := make(chan interface{}, 1024)
	writeChannel = c
	e := json.NewEncoder(w)
	go func() {
		for i := range c {
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
	}()
	return
}
