package service

import astral "github.com/cryptopunkscc/astrald/mod/apphost/client"

func (srv *Context) Serve(handlers Handlers) {
	for query, handle := range handlers {
		// start handler for query
		go func(srv Context, query string, handle Handler) {
			// register port for handler
			srv.LogPrefix(">", query)
			port := srv.register(query)
			for request := range port.Next() {
				// handle request
				go handle(srv, request)
			}
		}(*srv, query, handle)
	}
}

func (srv *Context) register(query string) (port astral.Port) {
	port, err := srv.Register(query)
	if err != nil {
		srv.Panicln("Cannot register port", query, err)
	}
	go func() {
		<-srv.Done()
		_ = port.Close()
	}()
	return
}
