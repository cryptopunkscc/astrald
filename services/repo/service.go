package repo

import (
	"context"
	"github.com/cryptopunkscc/astrald/api"
	"github.com/cryptopunkscc/astrald/components/sio"
	"github.com/cryptopunkscc/astrald/components/storage/file"
	"github.com/cryptopunkscc/astrald/components/storage/repo"
	"github.com/cryptopunkscc/astrald/services"
	"github.com/cryptopunkscc/astrald/services/repo/handle"
	"github.com/cryptopunkscc/astrald/services/util/accept"
	"github.com/cryptopunkscc/astrald/services/util/auth"
	"github.com/cryptopunkscc/astrald/services/util/register"
	"github.com/cryptopunkscc/astrald/services/util/request"
	"log"
)

const Port = "repo"
const FilesPort = "files"

const (
	List    = 0
	Read    = 1
	Write   = 2
	Observe = 3
	Map     = 4
)

// =================== Constructors ====================

func NewRepoService() *Context {
	return NewService(Port,
		auth.Local,
		Handlers{
			List:    handle.List,
			Read:    handle.Read,
			Write:   handle.Write,
			Observe: handle.Observe,
			Map:     handle.Map,
		},
	)
}

func NewFilesService() *Context {
	return NewService(FilesPort,
		auth.All,
		Handlers{
			List: handle.List,
			Read: handle.Read,
		},
	)
}

func NewService(
	port string,
	authorize auth.Authorize,
	handlers Handlers,
) *Context {
	return &Context{
		Context: request.Context{
			Port:      port,
			Observers: map[sio.ReadWriteCloser]struct{}{},
		},
		Authorize: authorize,
		Handlers:  handlers,
	}
}

// =================== Context ====================

type Context struct {
	request.Context
	auth.Authorize
	Handlers
}

type Handle func(c *handle.Request)

type Handlers map[byte]Handle

// =================== Runner ====================

func (srv *Context) Run(ctx context.Context, core api.Core) error {
	repository := repo.NewAdapter(file.NewStorage(services.AstralHome))
	handler, err := register.Port(ctx, core, srv.Port)
	if err != nil {
		return err
	}

	for r := range handler.Requests() {
		if !srv.Authorize(core, r) {
			log.Println(srv.Port, "rejected remote connection")
			continue
		}

		req := &handle.Request{
			Context:                srv.Context,
			ReadWriteMapRepository: repository,
		}
		req.ReadWriteCloser = accept.Request(ctx, r)
		log.Println(srv.Port, "accepted connection")

		go func() {
			defer func() { _ = req.Close() }()

			var err error
			var requestType byte
			var _handle Handle

			if requestType, err = req.ReadByte(); err != nil {
				log.Println(req.Port, "error reading type", err)
				return
			}
			log.Println(req.Port, "received request type", requestType)

			if _handle = srv.Handlers[requestType]; _handle == nil {
				log.Println(req.Port, "unknown request type", requestType)
				return
			}

			log.Println(req.Port, "handle request type", requestType)
			_handle(req)
		}()
	}
	return nil
}
