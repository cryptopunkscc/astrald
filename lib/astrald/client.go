package astrald

import (
	"strings"

	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/channel"
	"github.com/cryptopunkscc/astrald/lib/apphost"
	"github.com/cryptopunkscc/astrald/lib/query"
)

type Client struct {
	config Config
}

var defaultClient *Client

func NewClient(config Config) *Client {
	client := Client{config: config}
	return &client
}

func (client *Client) Query(target string, method string, args any) (_ *apphost.Conn, err error) {
	var targetID *astral.Identity
	if target != "localnode" && target != "" {
		targetID, err = Dir().ResolveIdentity(target)
		if err != nil {
			return nil, err
		}
	}

	return client.RouteQuery(query.New(nil, targetID, method, args))
}

func (client *Client) QueryChannel(target string, method string, args any, cfg ...channel.ConfigFunc) (*channel.Channel, error) {
	conn, err := client.Query(target, method, args)
	if err != nil {
		return nil, err
	}
	return channel.New(conn, cfg...), nil
}

func (client *Client) RouteQuery(query *astral.Query) (*apphost.Conn, error) {
	// connect to the host
	host, err := client.connect()
	if err != nil {
		return nil, err
	}

	return host.RouteQuery(query)
}

func (client *Client) Listen() (*Listener, error) {
	l, err := NewListener(client.protocol())
	if err != nil {
		return nil, err
	}

	host, err := client.connect()
	if err != nil {
		l.Close()
		return nil, err
	}

	token := astral.NewNonce()

	l.SetToken(token)
	err = host.Register(host.GuestID(), l.String(), token)
	if err != nil {
		l.Close()
		return nil, err
	}

	go func() {
		defer l.Close()

		// wait for the session to end
		for {
			_, err = host.Receive()
			if err != nil {
				break
			}
		}
	}()

	go func() {
		<-l.Done()
		host.Close()
	}()

	return l, nil
}

// connect establishes a new connection to the host
func (client *Client) connect() (host *apphost.Host, err error) {
	host, err = apphost.Connect(client.config.Endpoint)
	if err != nil {
		return nil, err
	}

	if len(client.config.Token) > 0 {
		err = host.AuthToken(client.config.Token)
		if err != nil {
			return nil, err
		}
	}

	return host, nil
}

func (client *Client) protocol() string {
	return strings.SplitN(client.config.Endpoint, ":", 2)[0]
}

func DefaultClient() *Client {
	if defaultClient == nil {
		defaultClient = NewClient(DefaultConfig())
	}
	return defaultClient
}

func Query(target string, method string, args any) (*apphost.Conn, error) {
	return DefaultClient().Query(target, method, args)
}

func RouteQuery(query *astral.Query) (*apphost.Conn, error) {
	return DefaultClient().RouteQuery(query)
}

func QueryChannel(target string, method string, args any, cfg ...channel.ConfigFunc) (*channel.Channel, error) {
	return DefaultClient().QueryChannel(target, method, args, cfg...)
}

func Listen() (*Listener, error) {
	return DefaultClient().Listen()
}
