package tor

import (
	"context"
	"github.com/cryptopunkscc/astrald/infra"
	"log"
	"strings"
)

func (tor *Tor) Listen(ctx context.Context) (<-chan infra.Conn, error) {
	output := make(chan infra.Conn)

	// process incoming connections
	go func() {
		defer close(output)

		key, err := tor.getPrivateKey()
		if err != nil {
			log.Println("[tor] key error:", err)
			return
		}

		l, err := tor.backend.Listen(ctx, key)
		if err != nil {
			log.Println("[tor] listen error:", err)
			return
		}
		defer l.Close()

		tor.serviceAddr, _ = Parse(l.Addr())

		log.Println("[tor] listen", tor.serviceAddr.String())

		for {
			conn, err := l.Accept()
			switch {
			case err == nil:
			case strings.Contains(err.Error(), "use of closed network connection"):
				return
			default:
				log.Println("[tor] accept error:", err)
				return
			}
			output <- newConn(conn, Addr{}, false)
		}
	}()

	return output, nil
}
