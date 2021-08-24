package push

import (
	"context"
	"github.com/cryptopunkscc/astrald/api"
	"github.com/cryptopunkscc/astrald/components/sio"
	"github.com/cryptopunkscc/astrald/services/push/handle"
	"github.com/cryptopunkscc/astrald/services/sync"
	"github.com/cryptopunkscc/astrald/services/util/accept"
	"github.com/cryptopunkscc/astrald/services/util/register"
	"github.com/cryptopunkscc/astrald/services/util/request"
	"log"
)

const Port = "push"

const (
	Push    = 1
	Observe = 2
)

// =================== Constructors ====================

func NewService() *Context {
	return &Context{
		Port: Port,
		Handlers: Handlers{
			Push:    handle.Push,
			Observe: handle.Observe,
		},
	}
}

// =================== Context ====================

type Context struct {
	Port string
	Handlers Handlers
}

type Handle func(r *handle.Request) error

type Handlers map[byte]Handle

// =================== Runner ====================

func (srv *Context) Run(ctx context.Context, core api.Core) (err error) {
	var handler api.PortHandler
	if handler, err = register.Port(ctx, core, srv.Port); err != nil {
		return
	}

	observers := map[sio.ReadWriteCloser]struct{}{}
	syncClient := sync.NewClient(ctx, core)

	for conn := range handler.Requests() {
		r := &handle.Request{
			Context:   request.Context{
				Port: srv.Port,
				ReadWriteCloser: accept.Request(ctx, conn),
				Observers: observers,
			},
			Caller:    conn.Caller(),
			Sync:      syncClient,
		}
		log.Println(srv.Port, "accepted connection")
		go func() {
			defer func() { _ = r.Close() }()
			var err error
			var requestType byte

			if requestType, err = r.ReadByte(); err != nil {
				log.Println(srv.Port, "cannot read request type", err)
				return
			}

			log.Println(srv.Port, "getting handler for request type", requestType)
			h := srv.Handlers[requestType]
			if h == nil {
				log.Println(srv.Port, "cannot obtain handler for request type", requestType, "len", len(srv.Handlers), srv.Handlers, err)
				return
			}
			log.Println(srv.Port, "handling request type", requestType)
			err = h(r)
			if err != nil {
				log.Println(srv.Port, "cannot handle request type", requestType, err)
			}
		}()
	}
	return
}
