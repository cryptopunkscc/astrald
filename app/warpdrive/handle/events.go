package handle

import (
	"encoding/json"
	"github.com/cryptopunkscc/astrald/app/warpdrive/api"
	"github.com/cryptopunkscc/astrald/app/warpdrive/service"
	"github.com/cryptopunkscc/astrald/enc"
	astral "github.com/cryptopunkscc/astrald/mod/apphost/client"
	"io"
	"sync"
)

func (r recipient) Events() (incoming <-chan api.Status, err error) {
	// Connect to local service
	conn, err := r.query(api.RecEvents)
	if err != nil {
		return
	}
	inc := make(chan api.Status)
	incoming = inc
	go func(conn io.ReadWriteCloser, inc chan api.Status) {
		defer close(inc)
		defer conn.Close()
		dec := json.NewDecoder(conn)
		files := &api.Status{}
		r.Println("Start listening status")
		for {
			err := dec.Decode(files)
			if err != nil {
				r.Println("Cannot decode status", err)
				return
			}
			inc <- *files
		}
	}(conn, inc)
	return
}

func RecipientEvents(srv service.Context, request astral.Request) {
	if srv.IsRejected(request) {
		return
	}
	// Accept connection
	conn, err := request.Accept()
	if err != nil {
		srv.Println("Cannot accept request", err)
		return
	}
	defer conn.Close()
	remove := srv.IncomingStatus().Subscribe(conn)
	defer remove()
	// Wait for close
	_, _ = enc.ReadUint8(conn)
}

func (s sender) Events() (outgoing <-chan api.Status, err error) {
	// Connect to local service
	conn, err := s.query(api.SenEvents)
	if err != nil {
		return
	}
	out := make(chan api.Status)
	outgoing = out
	go func(conn io.ReadWriteCloser, inc chan api.Status) {
		defer close(inc)
		defer conn.Close()
		dec := json.NewDecoder(conn)
		files := &api.Status{}
		s.Println("Start listening status")
		for {
			err := dec.Decode(files)
			if err != nil {
				s.Println("Finish listening offers status", err)
				return
			}
			inc <- *files
		}
	}(conn, out)
	return
}

func SenderEvents(srv service.Context, request astral.Request) {
	if srv.IsRejected(request) {
		return
	}
	// Accept connection
	conn, err := request.Accept()
	if err != nil {
		srv.Println("Cannot accept request", err)
		return
	}
	defer conn.Close()
	remove := srv.OutgoingStatus().Subscribe(conn)
	defer remove()
	// Wait for close
	_, _ = enc.ReadUint8(conn)
}

func (s *apiClient) Events() (events <-chan api.Status, err error) {
	senderEvents, err := s.sender.Events()
	if err != nil {
		return
	}
	recipientEvents, err := s.recipient.Events()
	if err != nil {
		return
	}
	events = merge(senderEvents, recipientEvents)
	return
}

func merge(cs ...<-chan api.Status) <-chan api.Status {
	out := make(chan api.Status)
	var wg sync.WaitGroup
	wg.Add(len(cs))
	for _, c := range cs {
		go func(c <-chan api.Status) {
			for v := range c {
				out <- v
			}
			wg.Done()
		}(c)
	}
	go func() {
		wg.Wait()
		close(out)
	}()
	return out
}
