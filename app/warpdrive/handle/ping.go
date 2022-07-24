package handle

import (
	"github.com/cryptopunkscc/astrald/app/warpdrive/handler"
	"github.com/cryptopunkscc/astrald/cslq"
	"github.com/cryptopunkscc/astrald/lib/astral"
)

func Ping(srv handler.Context, request astral.Request) {
	if srv.IsRejected(request) {
		return
	}
	// Accept connection
	conn, err := request.Accept()
	if err != nil {
		srv.Println("Cannot accept request", err)
		return
	}
	defer conn.Close()
	for {
		var code byte
		err = cslq.Decode(conn, "c", &code)
		srv.Println("Read ping", err)
		if err != nil {
			srv.Println("Cannot read ping", err)
			srv.Println("Closing connection")
			return
		}
	}
}
