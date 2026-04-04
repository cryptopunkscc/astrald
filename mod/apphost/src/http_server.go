package apphost

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/lib/ipc"
	"github.com/cryptopunkscc/astrald/mod/objects/fs"
)

const HTTPAuthTokenHeader = "X-Astral-Auth-Token"
const HTTPTargetHeader = "X-Astral-Target"

type HTTPServer struct {
	*Module
	fileSystem *fs.FS
	fileServer http.Handler
	ctx        *astral.Context
}

func NewHTTPServer(mod *Module) *HTTPServer {
	srv := &HTTPServer{
		Module: mod,
	}

	srv.fileSystem = fs.NewFS(srv.Objects.ReadDefault())
	srv.fileServer = http.FileServer(http.FS(srv.fileSystem))

	return srv
}

func (srv *HTTPServer) Run(ctx *astral.Context) error {
	// return if not configured
	bindHTTP := srv.config.BindHTTP
	if bindHTTP == "" {
		return nil
	}

	// bind to the address
	l, err := ipc.Listen(bindHTTP)
	if err != nil {
		srv.log.Error("http server: bind %v: %v", bindHTTP, err)
		return err
	}

	httpServer := &http.Server{Handler: srv}

	go func() {
		<-ctx.Done()
		sctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()

		err := httpServer.Shutdown(sctx)
		if err != nil {
			srv.log.Error("http server: shutdown err:: %v", bindHTTP, err)
		} else {
			srv.log.Logv(1, "http server: stopped", bindHTTP)
		}
	}()

	// start the server
	srv.log.Infov(1, "http server: started at %v", bindHTTP)
	err = httpServer.Serve(l)
	switch {
	case err == nil:
	case errors.Is(err, http.ErrServerClosed):
	default:
		srv.log.Error("http server: serve: %v", bindHTTP, err)
	}

	<-ctx.Done()
	return nil
}

func (srv *HTTPServer) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	var authToken = srv.getAuthToken(request)

	clientID, err := srv.AuthenticateToken(authToken)
	if err != nil {
		writer.WriteHeader(http.StatusUnauthorized)
		return
	}

	writer.Header().Set("X-Astral-Guest-Identity", clientID.String())
	writer.Header().Set("X-Astral-Host-Identity", srv.node.Identity().String())

	// handle object requests
	if after, found := strings.CutPrefix(request.URL.Path, "/.objects/"); found {
		// rewrite the path
		request.URL.Path = "/" + after
		NewHTTPObjectHandler(srv.Module, clientID).ServeHTTP(writer, request)
		return
	}

	NewHTTPQueryHandler(srv.Module, clientID).ServeHTTP(writer, request)
}

func (src *HTTPServer) getAuthToken(request *http.Request) (token string) {
	_, token, _ = strings.Cut(request.Header.Get("Authorization"), "Bearer ")
	return token
}
