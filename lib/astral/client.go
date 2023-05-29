package astral

import (
	"errors"
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/mod/apphost/proto"
	"os"
	"strings"
)

const defaultApphostAddr = "tcp:127.0.0.1:8625"

type ClientInfo struct {
	addr  string
	token string
}

func (c *ClientInfo) Storage(identity id.Identity, service string) *Storage {
	return &Storage{
		client:   c,
		Identity: identity,
		Service:  service,
	}
}

func (c *ClientInfo) LocalStorage() *Storage {
	return c.Storage(id.Identity{}, "storage.read")
}

var Client ClientInfo

func NewClient(addr string, token string) *ClientInfo {
	return &ClientInfo{
		addr:  addr,
		token: token,
	}
}

func (c *ClientInfo) Session() (*Session, error) {
	if len(c.addr) == 0 {
		return nil, errors.New("missing apphost address")
	}

	conn, err := proto.Dial(c.addr)
	if err == nil {
		return NewSession(conn, c.token, c.addr), nil
	}

	return nil, errors.New("apphost unrachable")
}

func (c *ClientInfo) Query(remoteID id.Identity, query string) (conn *Conn, err error) {
	s, err := c.Session()
	if err != nil {
		return nil, err
	}

	return s.Query(remoteID, query)
}

func (c *ClientInfo) QueryName(name string, query string) (conn *Conn, err error) {
	identity, err := c.Resolve(name)
	if err != nil {
		return
	}

	return c.Query(identity, query)
}

func (c *ClientInfo) Resolve(name string) (id.Identity, error) {
	s, err := c.Session()
	if err != nil {
		return id.Identity{}, err
	}

	return s.Resolve(name)
}

func (c *ClientInfo) NodeInfo(identity id.Identity) (info proto.NodeInfoData, err error) {
	s, err := c.Session()
	if err != nil {
		return
	}

	return s.NodeInfo(identity)
}

func (c *ClientInfo) Register(service string) (l *Listener, err error) {
	s, err := c.Session()
	if err != nil {
		return
	}

	l, err = newListener(s.proto())
	if err != nil {
		return
	}

	err = s.Register(service, l.Target())
	if err != nil {
		l.Close()
		return
	}

	l.onClose = func() {
		s.Close()
	}

	go func() {
		defer l.Close()
		var buf [16]byte
		for {
			if _, err := s.conn.Read(buf[:]); err != nil {
				return
			}
		}
	}()

	return
}

func (c *ClientInfo) Exec(identity id.Identity, app string, args []string, env []string) error {
	s, err := c.Session()
	if err != nil {
		return err
	}

	return s.Exec(identity, app, args, env)
}

func Exec(identity id.Identity, app string, args []string, env []string) error {
	return Client.Exec(identity, app, args, env)
}

func Query(remoteID id.Identity, query string) (*Conn, error) {
	return Client.Query(remoteID, query)
}

func QueryName(name string, query string) (conn *Conn, err error) {
	return Client.QueryName(name, query)
}

func Resolve(name string) (id.Identity, error) {
	return Client.Resolve(name)
}

func GetNodeInfo(identity id.Identity) (info proto.NodeInfoData, err error) {
	return Client.NodeInfo(identity)
}

func Register(service string) (l *Listener, err error) {
	return Client.Register(service)
}

func init() {
	var addrs []string
	var envAddr = os.Getenv(proto.EnvKeyAddr)

	if len(envAddr) > 0 {
		addrs = strings.Split(envAddr, ";")
	} else {
		addrs = []string{defaultApphostAddr}
	}

	for _, addr := range addrs {
		conn, err := proto.Dial(addr)
		if err == nil {
			conn.Close()
			Client.addr = addr
			break
		}
	}

	Client.token = os.Getenv(proto.EnvKeyToken)
}
