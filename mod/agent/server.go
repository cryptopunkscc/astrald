package agent

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/cryptopunkscc/astrald/log"
	"github.com/cryptopunkscc/astrald/node"
	"github.com/cryptopunkscc/astrald/node/services"
	"io"
	"os"
	"reflect"
	"time"
)

const requestTimeout = 30 * time.Second
const envAuthCookie = "ASTRALD_AGENT_COOKIE"

const (
	mGetAlias = "get_alias"
	mSetAlias = "set_alias"
	mAuth     = "auth"
)

var ErrRequestTimedOut = errors.New("request timed out")

// API errors
var (
	ErrInvalidMethod  = errors.New("invalid_method")
	ErrAuthFailed     = errors.New("auth_failed")
	ErrUnauthorized   = errors.New("unauthorized")
	ErrInvalidRequest = errors.New("invalid_request")
	ErrUnsupported    = errors.New("unsupported")
	ErrInternalError  = errors.New("internal_error")
)

type Server struct {
	node node.Node
	log  *log.Logger
	conn *services.Conn
	auth bool
}

type rawRequest struct {
	Method string          `json:"method"`
	Params json.RawMessage `json:"params"`
}

type GetAliasRequest struct {
}

type SetAliasRequest struct {
	Alias string `json:"alias"`
}

type AuthRequest struct {
	Cookie string `json:"cookie"`
}

type response struct {
	Status string      `json:"status"`
	Data   interface{} `json:"data"`
	Error  string      `json:"error"`
}

func (srv *Server) Run(ctx context.Context) error {
	defer srv.conn.Close()

	var requests = srv.readRequests()

	for {
		select {
		case req := <-requests:
			if err, ok := req.(error); ok {
				if !errors.Is(err, io.EOF) {
					srv.writeError(ErrInvalidRequest)
				}
				return err
			}

			srv.log.Infov(2, "request: %s", reflect.TypeOf(req).String())

			var res = srv.handleRequest(req)

			if err, ok := res.(error); ok {
				if err2 := srv.writeError(err); err2 != nil {
					return err2
				}
				continue
			}

			bytes, err := json.Marshal(&response{
				Status: "ok",
				Data:   res,
			})

			if err != nil {
				return err
			}

			srv.conn.Write(bytes)

		case <-ctx.Done():
			return ctx.Err()

		case <-time.After(requestTimeout):
			return ErrRequestTimedOut
		}
	}
}

func (srv *Server) writeError(e error) (err error) {
	bytes, err := json.Marshal(&response{
		Status: "error",
		Error:  e.Error(),
	})

	if err != nil {
		return
	}

	_, err = srv.conn.Write(bytes)
	return
}

func (srv *Server) handleRequest(req interface{}) interface{} {
	switch req := req.(type) {
	case AuthRequest:
		if req.Cookie == os.Getenv(envAuthCookie) {
			srv.auth = true
			return nil
		}

		return ErrAuthFailed

	case SetAliasRequest:
		if !srv.auth {
			return ErrUnauthorized
		}

		return srv.node.Tracker().SetAlias(srv.node.Identity(), req.Alias)

	case GetAliasRequest:
		if !srv.auth {
			return ErrUnauthorized
		}

		alias, err := srv.node.Tracker().GetAlias(srv.node.Identity())
		if err != nil {
			return ErrInternalError
		}

		return struct {
			Alias string `json:"alias"`
		}{
			Alias: alias,
		}
	}

	return ErrInvalidMethod
}

func (srv *Server) readRequests() <-chan interface{} {
	var jDec = json.NewDecoder(srv.conn)
	var out = make(chan interface{}, 1)

	go func() {
		defer close(out)
		for {
			var rr rawRequest
			var err error

			err = jDec.Decode(&rr)
			if err != nil {
				out <- err
				return
			}

			switch rr.Method {
			case mGetAlias:
				out <- GetAliasRequest{}

			case mSetAlias:
				var r SetAliasRequest

				if err := json.Unmarshal(rr.Params, &r); err != nil {
					out <- err
					return
				}

				out <- r

			case mAuth:
				var r AuthRequest

				if err := json.Unmarshal(rr.Params, &r); err != nil {
					out <- err
					return
				}

				out <- r

			default:
				out <- rr
			}
		}
	}()

	return out
}
