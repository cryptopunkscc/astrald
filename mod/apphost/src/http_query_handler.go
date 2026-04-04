package apphost

import (
	"errors"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/lib/query"
)

type HTTPQueryHandler struct {
	*Module
	Identity *astral.Identity
}

func NewHTTPQueryHandler(mod *Module, identity *astral.Identity) *HTTPQueryHandler {
	return &HTTPQueryHandler{Module: mod, Identity: identity}
}

func (srv *HTTPQueryHandler) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	var err error

	// add CORS headers
	writer.Header().Set("Access-Control-Allow-Origin", "*")
	writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
	writer.Header().Set("Access-Control-Allow-Headers", "*")

	// handle preflight OPTIONS request
	if request.Method == "OPTIONS" {
		writer.WriteHeader(http.StatusOK)
		return
	}

	// establish the target identity
	targetID := srv.node.Identity()

	target := request.Header.Get(HTTPTargetHeader)
	if target != "" {
		targetID, err = astral.ParseIdentity(target)
		if err != nil {
			writer.WriteHeader(http.StatusBadRequest)
			return
		}
	}

	// parse the query
	method, params := query.Parse(request.URL.String())
	method = strings.TrimPrefix(method, "/")

	// prepare the query context
	ctx, cancel := astral.
		NewContext(nil).
		WithZone(astral.ZoneAll).
		WithIdentity(srv.node.Identity()).
		WithTimeout(5 * time.Second)
	defer cancel()

	// route the query
	ch, err := query.RouteChan(
		ctx,
		srv.node,
		query.New(srv.Identity, targetID, method, params),
	)
	switch {
	case err == nil:
	case errors.Is(err, &astral.ErrRejected{}):
		writer.WriteHeader(http.StatusMethodNotAllowed)
		return
	default:
		writer.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer ch.Close()

	// forward the traffic
	writer.WriteHeader(http.StatusOK)
	_, err = io.Copy(writer, ch.Transport())
}
