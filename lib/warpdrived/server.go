package warpdrived

import (
	"context"
	"errors"
	"fmt"
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/lib/warpdrived/core"
	"github.com/cryptopunkscc/astrald/lib/warpdrived/service"
	"github.com/cryptopunkscc/astrald/lib/wrapper"
	"github.com/cryptopunkscc/astrald/proto/warpdrive"
	"log"
)

type Server struct {
	context.Context
	core.Component
	warp    warpdrive.Service
	localId id.Identity
	Debug   bool
	Finish  <-chan struct{}
}

func (s *Server) String() string {
	return "[warpdrive]"
}

func (s *Server) Run(ctx context.Context, api wrapper.Api) (err error) {

	s.Api = api
	s.Context = ctx

	if s.localId, err = s.Resolve("localnode"); err != nil {
		return errors.New(fmt.Sprintln("Cannot resolve local id;", err))
	}

	setupCore(&s.Component)

	s.Finish = service.OfferUpdates(s.Component).Start(s)

	s.warp = service.Warpdrive(s.Component)

	s.warp.Peer().Fetch()

	if err = s.register(warpdrive.Port, warpdrive.Dispatch); err != nil {
		return warpdrive.Error(err, "Cannot register", warpdrive.Port)
	}

	if err = s.register(warpdrive.CliPort, warpdrive.Cli); err != nil {
		return warpdrive.Error(err, "Cannot register", warpdrive.CliPort)
	}

	<-s.Done()
	<-s.Finish
	return
}

func (s *Server) register(
	query string,
	dispatch func(d *warpdrive.Dispatcher) error,
) (err error) {
	port, err := s.Api.Register(query)
	if err != nil {
		return
	}

	// Serve handlers
	go func() {
		requestId := uint(0)
		for request := range port.Next() {
			requestId = requestId + 1
			go func(request wrapper.Request, requestId uint) {
				s.Job.Add(1)
				defer s.Job.Done()

				conn, err := request.Accept()
				defer func() {
					if err != nil {
						log.Println(err)
					}
				}()
				if err != nil {
					err = warpdrive.Error(err, "Cannot accept warpdrive connection")
					return
				}
				defer conn.Close()
				callerId := request.Caller().String()
				authorized := callerId == s.localId.String()
				_ = warpdrive.Dispatcher{
					Context:    s.Context,
					Service:    s.warp,
					CallerId:   callerId,
					Api:        s.Api,
					Job:        s.Job,
					Conn:       conn,
					Authorized: s.Debug || authorized,
					LogPrefix:  fmt.Sprint("[WARPDRIVE] ", query, ":", requestId),
				}.Serve(dispatch)
			}(request, requestId)
		}
	}()
	go func() {
		<-s.Done()
		port.Close()
	}()
	return
}
