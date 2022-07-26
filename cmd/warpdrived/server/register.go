package server

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
				_ = warpdrive.Dispatcher{
					Service:   s.warp,
					CallerId:  request.Caller().String(),
					LocalId:   s.localId.String(),
					Api:       s.Api,
					Conn:      conn,
					LogPrefix: fmt.Sprint("[WARPDRIVE] ", query, ":", requestId),
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
