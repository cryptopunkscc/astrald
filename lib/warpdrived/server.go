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
	core.Component
	ctx     context.Context
	warp    warpdrive.Service
	localId id.Identity
	debug   bool
}

func (s *Server) String() string {
	return "[warpdrive]"
}

func (s *Server) Run(ctx context.Context, api wrapper.Api) (err error) {

	s.ctx = ctx
	s.Api = api

	if s.localId, err = s.Resolve("localnode"); err != nil {
		return errors.New(fmt.Sprintln("Cannot resolve local id;", err))
	}

	setupCore(&s.Component)

	finish := service.OfferUpdates(s.Component).Start(s.ctx)

	s.warp = service.Warpdrive(s.Component)

	s.warp.Peer().Fetch()

	dispatchers := map[string]func(d *warpdrive.Dispatcher) error{
		warpdrive.PortInfo:   warpdrive.DispatchInfo,
		warpdrive.PortLocal:  warpdrive.DispatchLocal,
		warpdrive.PortRemote: warpdrive.DispatchRemote,
		warpdrive.PortCli:    warpdrive.Cli,
	}

	for port, dispatcher := range dispatchers {
		if err = s.register(port, dispatcher); err != nil {
			return warpdrive.Error(err, "Cannot register", port)
		}
	}

	<-s.ctx.Done()
	<-finish
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
				defer conn.Close()

				if err != nil {
					err = warpdrive.Error(err, "Cannot accept warpdrive connection")
					log.Println(err)
					return
				}
				logPrefix := fmt.Sprint("[WARPDRIVE] ", query, ":", requestId)
				callerId := request.Caller().String()
				authorized := callerId == s.localId.String()
				_ = warpdrive.NewDispatcher(
					logPrefix,
					callerId,
					s.debug || authorized,
					s.ctx,
					s.Api,
					conn,
					s.warp,
					s.Job,
				).Serve(dispatch)
			}(request, requestId)
		}
	}()
	go func() {
		<-s.ctx.Done()
		port.Close()
	}()
	return
}
