package astral

import (
	"github.com/cryptopunkscc/astrald/mod/apphost/proto"
	"io"
	"net"
)

type Port struct {
	path     string
	closer   io.Closer
	requests chan *Request
	listener net.Listener
}

func NewUnixPort(path string, closer io.Closer) (*Port, error) {
	var err error
	var port = &Port{
		path:     path,
		closer:   closer,
		requests: make(chan *Request),
	}

	port.listener, err = net.Listen("unix", port.path)
	if err != nil {
		return nil, err
	}

	go func() {
		defer closer.Close()
		defer close(port.requests)

		for {
			rawConn, err := port.listener.Accept()
			if err != nil {
				return
			}

			conn := proto.NewConn(rawConn)
			r, err := conn.ReadRequest()
			if err != nil {
				rawConn.Close()
				continue
			}

			port.requests <- &Request{
				caller: r.Identity,
				query:  r.Port,
				conn:   conn,
				raw:    rawConn,
			}
		}
	}()

	return port, nil
}

func (port *Port) Next() <-chan *Request {
	return port.requests
}

func (port *Port) Close() error {
	return port.listener.Close()
}
