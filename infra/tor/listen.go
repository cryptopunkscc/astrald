package tor

import (
	"context"
	"github.com/cryptopunkscc/astrald/infra"
	"io/ioutil"
	"log"
	"net"
	"os"
	"path"
)

func (tor *Tor) Listen(ctx context.Context) (<-chan infra.Conn, <-chan error) {
	output, errCh := make(chan infra.Conn), make(chan error)

	ctl, err := Open(tor.config.getContolAddr())
	if err != nil {
		return nil, singleErrCh(err)
	}

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return nil, singleErrCh(err)
	}

	key := ""
	newKey := true

	bytes, err := ioutil.ReadFile(keyPath())
	if err == nil {
		key = string(bytes)
		newKey = false
	}

	srv, err := ctl.StartService(
		key,
		map[int]string{
			tor.config.getListenPort(): listener.Addr().String(),
		},
	)
	if err != nil {
		return nil, singleErrCh(err)
	}

	tor.serviceAddr, _ = Parse(srv.serviceID)

	log.Printf("listen tor %s.onion:%d\n", srv.serviceID, defaultListenPort)
	if newKey {
		ioutil.WriteFile(keyPath(), []byte(srv.PrivateKey()), 0600)
	}

	go func() {
		defer close(output)

		go func() {
			<-ctx.Done()
			listener.Close()
		}()

		for {
			conn, err := listener.Accept()
			if err != nil {
				break
			}
			output <- newConn(conn, Addr{}, false)
		}

		log.Printf("closed tor %s.onion:%d\n", srv.serviceID, defaultListenPort)

		srv.Close()
	}()

	return output, errCh
}

// TODO: figure out how to do driver cache nicely
func keyPath() string {
	home, _ := os.UserHomeDir()
	return path.Join(home, ".config", "astrald", "tor.key")
}
