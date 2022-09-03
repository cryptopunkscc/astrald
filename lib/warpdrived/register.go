package warpdrived

import (
	"fmt"
	"github.com/cryptopunkscc/astrald/lib/wrapper"
	"github.com/cryptopunkscc/astrald/proto/warpdrive"
	"log"
)

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
