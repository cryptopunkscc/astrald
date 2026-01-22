package dir

import (
	"fmt"

	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/channel"
	"github.com/cryptopunkscc/astrald/lib/astrald"
	"github.com/cryptopunkscc/astrald/lib/query"
	"github.com/cryptopunkscc/astrald/mod/apphost"
	"github.com/cryptopunkscc/astrald/mod/dir"
	"github.com/cryptopunkscc/astrald/sig"
)

type Client struct {
	targetID     *astral.Identity
	astral       *astrald.Client
	resolveCache sig.Map[string, *astral.Identity]
	aliasCache   sig.Map[string, string]
}

var defaultClient *Client

func New(targetID *astral.Identity, client *astrald.Client) *Client {
	if client == nil {
		client = astrald.Default()
	}

	return &Client{
		astral:   client,
		targetID: targetID,
	}
}

func Default() *Client {
	if defaultClient == nil {
		defaultClient = New(nil, astrald.Default())
	}

	return defaultClient
}

func (client *Client) AliasMap(ctx *astral.Context) (*dir.AliasMap, error) {
	// query
	ch, err := client.queryCh(ctx, "dir.alias_map", nil)
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

func (client *Client) GetAlias(ctx *astral.Context, identity *astral.Identity) (string, error) {
	if alias, ok := client.aliasCache.Get(identity.String()); ok {
		return alias, nil
	}

	ch, err := client.queryCh(ctx, "dir.get_alias", query.Args{
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

func GetAlias(ctx *astral.Context, identity *astral.Identity) (string, error) {
	return Default().GetAlias(ctx, identity)
}

func (client *Client) ResolveIdentity(ctx *astral.Context, name string) (*astral.Identity, error) {
	// try to parse the public key first
	if id, err := astral.IdentityFromString(name); err == nil {
		return id, nil
	}

	// check cache
	if id, ok := client.resolveCache.Get(name); ok {
		return id, nil
	}

	// then try using host's resolver
	ch, err := client.queryCh(ctx, "dir.resolve", query.Args{"name": name})
	if err != nil {
		return nil, err
	}
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

func ResolveIdentity(ctx *astral.Context, name string) (*astral.Identity, error) {
	return Default().ResolveIdentity(ctx, name)
}

func (client *Client) ClearCache() {
	client.resolveCache = sig.Map[string, *astral.Identity]{}
	client.aliasCache = sig.Map[string, string]{}
}

func (client *Client) queryCh(ctx *astral.Context, method string, args any) (*channel.Channel, error) {
	return client.astral.WithTarget(client.targetID).QueryChannel(ctx, method, args)
}
