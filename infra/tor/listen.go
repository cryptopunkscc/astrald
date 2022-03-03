package tor

import (
	"context"
	"github.com/cryptopunkscc/astrald/infra"
	"log"
	"time"
)

const retryInterval = 15 * time.Second
const pingInterval = time.Minute

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
			if err != nil {
				log.Println("[tor] accept error:", err)
				return
			}
			output <- newConn(conn, Addr{}, false)
		}
	}()

	return output, nil
}
