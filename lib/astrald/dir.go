package astrald

import (
	"fmt"

	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/channel"
	"github.com/cryptopunkscc/astrald/lib/query"
	"github.com/cryptopunkscc/astrald/mod/apphost"
	"github.com/cryptopunkscc/astrald/mod/dir"
	"github.com/cryptopunkscc/astrald/sig"
)

type DirClient struct {
	astral       *Client
	resolveCache sig.Map[string, *astral.Identity]
	aliasCache   sig.Map[string, string]
}

var defaultDirClient *DirClient

func NewDirClient(c *Client) *DirClient {
	return &DirClient{astral: c}
}

func Dir() *DirClient {
	if defaultDirClient == nil {
		defaultDirClient = NewDirClient(DefaultClient())
	}
	return defaultDirClient
}

func (client *DirClient) ResolveIdentity(name string) (*astral.Identity, error) {
	// try to parse the public key first
	if id, err := astral.IdentityFromString(name); err == nil {
		return id, nil
	}

	// check cache
	if id, ok := client.resolveCache.Get(name); ok {
		return id, nil
	}

	// then try using host's resolver
	conn, err := client.astral.RouteQuery(query.New(nil, nil, "dir.resolve", query.Args{
		"name": name,
	}))
	if err != nil {
		return nil, err
	}

	ch := channel.New(conn)
	defer ch.Close()

	obj, err := ch.Receive()
	if err != nil {
		return nil, err
	}

	id, ok := obj.(*astral.Identity)
	if !ok {
		return nil, fmt.Errorf("unexpected type: %s", obj.ObjectType())
	}

	// cache results
	client.resolveCache.Set(name, id)

	return id, nil
}

func (client *DirClient) GetAlias(identity *astral.Identity) (string, error) {
	if alias, ok := client.aliasCache.Get(identity.String()); ok {
		return alias, nil
	}

	ch, err := client.astral.QueryChannel("", "dir.get_alias", query.Args{
		"id": identity,
	})
	if err != nil {
		return "", err
	}

	o, err := ch.Receive()
	switch o := o.(type) {
	case nil:
		return "", err
	case *astral.String8:
		client.aliasCache.Set(identity.String(), string(*o))
		return o.String(), nil
	default:
		return "", fmt.Errorf("unexpected type: %s", o.ObjectType())
	}
}

func (client *DirClient) AliasMap() (*dir.AliasMap, error) {
	// query
	ch, err := client.astral.QueryChannel("", "dir.alias_map", nil)
	if err != nil {
		return nil, err
	}

	// response
	o, err := ch.Receive()
	switch o := o.(type) {
	case nil:
		return nil, err

	case *dir.AliasMap:
		return o, nil

	default:
		return nil, apphost.ErrProtocolError
	}
}

func (client *DirClient) ClearCache() {
	client.resolveCache = sig.Map[string, *astral.Identity]{}
	client.aliasCache = sig.Map[string, string]{}
}
