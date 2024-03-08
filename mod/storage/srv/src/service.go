package srv

import (
	"context"
	"encoding/json"
	"github.com/cryptopunkscc/astrald/mod/storage"
	"github.com/cryptopunkscc/astrald/mod/storage/srv"
	"github.com/cryptopunkscc/astrald/net"
	"github.com/cryptopunkscc/astrald/node"
	"io"
	"log"
	"strings"
)

type Service struct {
	storage.Module
	node      node.Node
	port      string
	unmarshal UnmarshalFunc
	encoder   GetEncoder
	handlers  Handlers[Context]
}

type GetEncoder func(writer io.Writer) Encoder

func NewService(module storage.Module, node node.Node) (s *Service) {
	s = &Service{
		port:      proto.Port,
		Module:    module,
		node:      node,
		unmarshal: json.Unmarshal,
		encoder:   func(writer io.Writer) Encoder { return json.NewEncoder(writer) },
		handlers:  make(Handlers[Context]),
	}
	h := s.handlers

	Bind(h, proto.ReadAllReq{}, readAll)
	Bind(h, proto.PutReq{}, put)
	Bind(h, proto.OpenReq{}, open)
	Bind(h, proto.CreateReq{}, create)
	Bind(h, proto.PurgeReq{}, purge)

	return
}

func (s *Service) Run(ctx context.Context) (err error) {
	routeAny := false
	for _, h := range s.handlers {
		if h.Hidden() {
			routeAny = true
			continue
		}

		route := s.port + h.Query()
		if err = s.node.LocalRouter().AddRoute(route, s); err != nil {
			return
		}
		defer s.node.LocalRouter().RemoveRoute(route)
	}

	if routeAny {
		route := s.port + "*"
		if err = s.node.LocalRouter().AddRoute(route, s); err != nil {
			return
		}
		defer s.node.LocalRouter().RemoveRoute(route)
	}

	<-ctx.Done()
	return nil
}

func (s *Service) RouteQuery(_ context.Context, query net.Query, caller net.SecureWriteCloser, _ net.Hints) (net.SecureWriteCloser, error) {
	c := query.Caller()
	q := query.Query()
	cmd := s.parseCmd(q)

	h, ok := s.handlers[cmd]
	if !ok {
		return net.Reject()
	}

	ok = s.node.Auth().Authorize(c, h.Action())
	if !ok {
		return net.Reject()
	}

	return net.Accept(query, caller, func(conn net.SecureConn) {
		ctx := Context{
			Module:   s.Module,
			Conn:     conn,
			RemoteID: c,
		}
		if err := h.handle(ctx, conn, s.unmarshal, s.encoder(conn), q); err != nil {
			log.Println(err)
		}
	})
}

func (s *Service) parseCmd(query string) (cmd string) {
	cmd = strings.TrimPrefix(query, s.port)
	cmd = strings.SplitN(cmd, "?", 2)[0]
	return
}
