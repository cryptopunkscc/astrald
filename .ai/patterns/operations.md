# Operation Patterns

Use this pattern for operations exposed through `astral-query`, `lib/routing`,
or module client packages.

## Handler

Invariants:

* Method names use `Op` + PascalCase.
* Operation names use snake_case.

```go
func (mod *Module) OpStreams(ctx *astral.Context, q *routing.IncomingQuery, args opStreamsArgs) error {
    ch := channel.New(q.AcceptRaw(), channel.WithOutputFormat(args.Out))
    defer ch.Close()

    for _, s := range mod.peers.links.Clone() {
        if err := ch.Send(s.Info()); err != nil {
            return ch.Send(astral.Err(err))
        }
    }
    return ch.Send(&astral.EOS{})
}
```

Source: `mod/nodes/src/op_streams.go`

## Args Struct

Define a struct for query string args.

```go
type opStreamsArgs struct {
    Out string         `query:"optional"`
    Key astral.String8 `query:"optional"`
}
```

Invariants:

* Args are parsed from the query string automatically.
* Optional args need the `query:"optional"` tag.

## Module Client

Every module with query ops has a `mod/<name>/client/` package. It mirrors the
module's `src/op_*.go` files.

```go
type Client struct {
    astral   *astrald.Client
    targetID *astral.Identity
}

var defaultClient *Client

func New(targetID *astral.Identity, a *astrald.Client) *Client {
    if a == nil {
        a = astrald.Default()
    }
    return &Client{astral: a, targetID: targetID}
}

func Default() *Client {
    if defaultClient == nil {
        defaultClient = New(nil, astrald.Default())
    }
    return defaultClient
}

func (c *Client) queryCh(ctx *astral.Context, method string, args any, cfg ...channel.ConfigFunc) (*channel.Channel, error) {
    return c.astral.WithTarget(c.targetID).QueryChannel(ctx, method, args, cfg...)
}
```

Source: `mod/nat/client/client.go`, `mod/objects/client/client.go`,
`mod/auth/client/client.go`

## Client Operation File

Keep one client operation per file.

```go
func (c *Client) DoThing(ctx *astral.Context, contract *foo.Contract) (result *foo.SignedContract, err error) {
    ch, err := c.queryCh(ctx, foo.OpDoThing, nil)
    if err != nil {
        return
    }
    defer ch.Close()

    err = ch.Send(contract)
    if err != nil {
        return
    }

    err = ch.Switch(channel.Expect(&result), channel.PassErrors)
    return
}

func DoThing(ctx *astral.Context, contract *foo.Contract) (*foo.SignedContract, error) {
    return Default().DoThing(ctx, contract)
}
```

Source: `mod/objects/client/store.go`, `mod/auth/client/sign_contract.go`

## Call Boundary

Choose the call path by caller situation.

| Situation | Use |
|---|---|
| Module running on same node | Dependency interface, for example `mod.Dir.ResolveIdentity(name)` |
| Operation on a different node | Client with target, for example `dirClient.New(target, astrald.Default())` |
| External app with no node access | Default client routed through apphost |

Source: `mod/nat/src/op_punch.go`, `mod/nat/src/op_node_consume_hole.go`
