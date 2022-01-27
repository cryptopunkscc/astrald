package handler

import astral "github.com/cryptopunkscc/astrald/mod/apphost/client"

func (ctx Context) Serve(handlers Handlers) {
	for query, handle := range handlers {
		go ctx.serve(query, handle)
	}
}

func (ctx Context) serve(query string, handle Handler) {
	ctx.LogPrefix(">", query)
	port := ctx.register(query)
	for request := range port.Next() {
		go handle(ctx, request)
	}
}

func (ctx *Context) register(query string) (port astral.Port) {
	port, err := ctx.Register(query)
	if err != nil {
		ctx.Panicln("Cannot register port", query, err)
	}
	go func() {
		<-ctx.Done()
		_ = port.Close()
	}()
	return
}
