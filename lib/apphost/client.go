package apphost

import (
	"errors"
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/lib/query"
	"github.com/cryptopunkscc/astrald/mod/apphost"
	"github.com/cryptopunkscc/astrald/sig"
	"os"
	"strings"
)

const DefaultEndpoint = "tcp:127.0.0.1:8625"
const AuthTokenEnv = "ASTRALD_APPHOST_TOKEN"

type Client struct {
	Endpoint      string
	AuthToken     string
	GuestID       *astral.Identity
	HostID        *astral.Identity
	resolverCache sig.Map[string, *astral.Identity]
}

func NewClient(endpoint string, token string) (*Client, error) {
	s, err := Connect(endpoint)
	if err != nil {
		return nil, err
	}
	defer s.Close()

	var res apphost.AuthResponse
	if len(token) == 0 {
		res, err = s.Anon()
	} else {
		res, err = s.Token(token)
	}
	if err != nil {
		return nil, err
	}

	if res.Code != apphost.Success {
		return nil, errors.New("authentication failed")
	}

	return &Client{
		Endpoint:  endpoint,
		AuthToken: token,
		GuestID:   res.GuestID,
		HostID:    res.HostID,
	}, nil
}

func NewDefaultClient() (*Client, error) {
	return NewClient(DefaultEndpoint, os.Getenv(AuthTokenEnv))
}

func (c *Client) Session() (*Session, error) {
	s, err := Connect(c.Endpoint)
	if err != nil {
		return nil, err
	}

	var res apphost.AuthResponse
	if len(c.AuthToken) == 0 {
		res, err = s.Anon()
	} else {
		res, err = s.Token(c.AuthToken)
	}
	if err != nil {
		return nil, err
	}

	if res.Code != apphost.Success {
		s.Close()
		return nil, errors.New("authentication failed")
	}

	return s, nil
}

func (c *Client) Query(target string, method string, args any) (*Conn, error) {
	targetID, err := c.resolveIdentityCached(target)
	if err != nil {
		return nil, err
	}

	var q = method

	if args != nil {
		params, err := query.Marshal(args)
		if err != nil {
			return nil, err
		}

		if len(params) > 0 {
			q = q + "?" + params
		}
	}

	s, err := c.Session()
	if err != nil {
		return nil, err
	}

	return s.Query(c.GuestID, targetID, q)
}

func (c *Client) Listen() (*Listener, error) {
	l, err := NewListener(c.Protocol())
	if err != nil {
		return nil, err
	}

	s, err := c.Session()
	if err != nil {
		l.Close()
		return nil, err
	}

	token, err := s.Register(c.GuestID, l.String())
	if err != nil {
		l.Close()
		return nil, err
	}

	go func() {
		var buf [1]byte
		s.conn.Read(buf[:])
		l.Close()
		s.conn.Close()
	}()

	go func() {
		<-l.Done()
		s.conn.Close()
	}()

	l.SetToken(token)
	return l, nil
}

func (c *Client) Protocol() string {
	return strings.SplitN(c.Endpoint, ":", 2)[0]
}

func (c *Client) Target(identity *astral.Identity) *Target {
	return &Target{targetID: identity, client: c}
}

func (c *Client) LocalNode() *Target {
	return c.Target(c.HostID)
}

func (c *Client) resolveIdentityCached(name string) (*astral.Identity, error) {
	// try to parse the public key first
	if id, err := astral.IdentityFromString(name); err == nil {
		return id, nil
	}

	// check cache
	if id, ok := c.resolverCache.Get(name); ok {
		return id, nil
	}

	// resolve the identity using the local node's resolver
	id, err := c.LocalNode().ResolveIdentity(name)
	if err != nil {
		return nil, err
	}

	// cache results
	c.resolverCache.Set(name, id)

	return id, nil
}
