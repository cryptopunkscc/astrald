package apphost

import (
	"context"
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/lib/ipc"
	"github.com/cryptopunkscc/astrald/mod/objects"
	"github.com/cryptopunkscc/astrald/mod/objects/fs"
	"github.com/cryptopunkscc/astrald/object"
	"net/http"
	"path/filepath"
	"sync"
	"time"
)

const HTTPAuthTokenHeader = "X-Astral-Auth-Token"
const HTTPAuthTokenParam = "auth_token"

type ObjectServer struct {
	*Module
	fileSystem *fs.FS
	fileServer http.Handler
	ctx        *astral.Context
}

func NewObjectServer(mod *Module) *ObjectServer {
	srv := &ObjectServer{
		Module: mod,
	}

	srv.fileServer = http.FileServer(http.FS(srv.fileSystem))

	return srv
}

func (srv *ObjectServer) Run(ctx *astral.Context) error {
	srv.ctx = ctx
	defer func() {
		srv.ctx = nil
	}()

	srv.fileSystem = fs.NewFS(ctx, srv.Deps.Objects)

	var wg sync.WaitGroup

	for _, bind := range srv.config.ObjectServer.Bind {
		bind := bind

		l, err := ipc.Listen(bind)
		if err != nil {
			srv.log.Error("object server failed to bind to %v: %v", bind, err)
			continue
		}

		// start the server
		srv.log.Infov(1, "object server started at %v", bind)
		httpServer := &http.Server{Handler: srv}
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := httpServer.Serve(l); err != nil && err != http.ErrServerClosed {
				srv.log.Error("object server at %v ended with error: %v", bind, err)
			}
		}()

		go func() {
			<-ctx.Done()
			sctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
			defer cancel()
			if err := httpServer.Shutdown(sctx); err != nil {
				srv.log.Error("object server shutdown at %v failed: %v", bind, err)
			} else {
				srv.log.Logv(1, "object server at %v stopped", bind)
			}
		}()
	}

	<-ctx.Done()
	wg.Wait()
	return nil
}

func (srv *ObjectServer) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	var authToken string

	authToken = request.URL.Query().Get(HTTPAuthTokenParam)

	if authToken == "" {
		authToken = request.Header.Get(HTTPAuthTokenHeader)
	}

	clientID := &astral.Identity{}
	token, err := srv.db.FindAccessToken(authToken)

	if token != nil {
		clientID = token.Identity
	}

	filename := filepath.Base(request.URL.Path)

	objectID, err := object.ParseID(filename)
	if err != nil {
		writer.WriteHeader(http.StatusNotFound)
		return
	}

	if !srv.Deps.Auth.Authorize(clientID, objects.ActionRead, objectID) {
		writer.WriteHeader(http.StatusForbidden)
		return
	}

	writer.Header().Set("Content-Disposition", "inline; filename="+objectID.String())

	ctx := srv.ctx.ExcludeZone(astral.ZoneNetwork)

	fserve := http.FileServer(http.FS(fs.NewFS(ctx, srv.Objects)))

	fserve.ServeHTTP(writer, request)
}
