package astral

import (
	"errors"
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/mod/apphost/proto"
	"math/rand"
	"os"
	"strings"
)

const defaultApphostAddr = "tcp:127.0.0.1:8625"

type ApphostClient struct {
	addr  string
	token string
}

var Client ApphostClient

func NewClient(addr string, token string) *ApphostClient {
	return &ApphostClient{
		addr:  addr,
		token: token,
	}
}

func (c *ApphostClient) Session() (*Session, error) {
	if len(c.addr) == 0 {
		return nil, errors.New("missing apphost address")
	}

	conn, err := proto.Dial(c.addr)
	if err == nil {
		return NewSession(conn, c.token, c.addr), nil
	}

	return nil, errors.New("apphost unrachable")
}

func (c *ApphostClient) Query(remoteID *astral.Identity, query string) (conn *Conn, err error) {
	s, err := c.Session()
	if err != nil {
		return nil, err
	}

	return s.Query(remoteID, query)
}

func (c *ApphostClient) QueryName(name string, query string) (conn *Conn, err error) {
	identity, err := c.Resolve(name)
	if err != nil {
		return
	}

	return c.Query(identity, query)
}

func (c *ApphostClient) Resolve(name string) (*astral.Identity, error) {
	s, err := c.Session()
	if err != nil {
		return nil, err
	}

	return s.Resolve(name)
}

func (c *ApphostClient) NodeInfo(identity *astral.Identity) (info proto.NodeInfoData, err error) {
	s, err := c.Session()
	if err != nil {
		return
	}

	return s.NodeInfo(identity)
}

func (c *ApphostClient) Register(service string) (l *Listener, err error) {
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

func (c *ApphostClient) Exec(identity *astral.Identity, app string, args []string, env []string) error {
	s, err := c.Session()
	if err != nil {
		return err
	}

	return s.Exec(identity, app, args, env)
}

func Exec(identity *astral.Identity, app string, args []string, env []string) error {
	return Client.Exec(identity, app, args, env)
}

func Query(remoteID *astral.Identity, query string) (*Conn, error) {
	return Client.Query(remoteID, query)
}

func QueryName(name string, query string) (conn *Conn, err error) {
	return Client.QueryName(name, query)
}

func Resolve(name string) (*astral.Identity, error) {
	return Client.Resolve(name)
}

func GetNodeInfo(identity *astral.Identity) (info proto.NodeInfoData, err error) {
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

func randomName(length int) (s string) {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	var name = make([]byte, length)
	for i := 0; i < len(name); i++ {
		name[i] = charset[rand.Intn(len(charset))]
	}
	return string(name[:])
}
