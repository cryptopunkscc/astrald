package server

import (
	"context"
	"errors"
	"fmt"
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/cmd/warpdrived/core"
	"github.com/cryptopunkscc/astrald/cmd/warpdrived/service"
	"github.com/cryptopunkscc/astrald/lib/wrapper"
	"github.com/cryptopunkscc/astrald/proto/warpdrive"
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
