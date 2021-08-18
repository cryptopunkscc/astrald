package tor

import (
	"context"
	"github.com/cryptopunkscc/astrald/node/net"
	"io/ioutil"
	"log"
	_net "net"
	"os"
	"path"
)

func (drv *driver) Listen(ctx context.Context) (<-chan net.Conn, error) {
	output := make(chan net.Conn)

	ctl, err := Open()
	if err != nil {
		return nil, err
	}

	listener, err := _net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return nil, err
	}

	key := ""
	newKey := true

	bytes, err := ioutil.ReadFile(keyPath())
	if err == nil {
		key = string(bytes)
		newKey = false
	}

	srv, err := ctl.StartService(key, map[int]string{1791: listener.Addr().String()})
	if err != nil {
		return nil, err
	}

	log.Printf("listen tor %s.onion:1791\n", srv.serviceID)
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
			output <- net.WrapConn(conn, false)
		}

		log.Printf("closed tor %s.onion:1791\n", srv.serviceID)

		srv.Close()
	}()

	return output, nil
}

// TODO: figure out how to do driver cache nicely
func keyPath() string {
	home, _ := os.UserHomeDir()
	return path.Join(home, ".config", "astrald", "tor.key")
}
