# Operation Patterns

Use this pattern for operations exposed through `astral-query`, `lib/routing`,
or module client packages.

## Handler

Invariants:

- Method names use `Op` + PascalCase (registered via `AddStructPrefix(s, "Op")`).
- Operation names use snake_case (auto-converted by `log.ToSnakeCase`).
- Handler signature: `func(*astral.Context, *routing.IncomingQuery[, args]) error`.

```go
func (mod *Module) OpSessions(ctx *astral.Context, q *routing.IncomingQuery, args opSessionsArgs) (err error) {
    ch := channel.New(q.AcceptRaw(), channel.WithOutputFormat(args.Out))
    defer ch.Close()

    for _, s := range mod.sessions() {
        if err = ch.Send(s); err != nil {
            return ch.Send(astral.NewError(err.Error()))
        }
    }
    return ch.Send(&astral.EOS{})
}
```

`q.AcceptRaw()` returns `io.ReadWriteCloser`; pass it to `channel.New(...)`.
`q.Accept(cfg...)` is the same flow in one call.

Source: `mod/nodes/src/op_sessions.go`, `mod/nodes/src/op_links.go`

## Args Struct

Args are parsed from the query string into the third handler argument. Fields
with `query:"required"` are enforced by `Op.invoke` before the handler runs;
otherwise the field is taken when present and left zero when absent.
`query:"optional"` is a documentation marker (it is parsed but defaults match
absence-of-tag).

```go
type opPunchArgs struct {
    Target string                  // present when set in query string
    In     string `query:"optional"`
    Out    string `query:"optional"`
}
```

For strictly required fields use `required`:

```go
type findArgs struct {
    ID  *astral.ObjectID `query:"required"`
    Out string           `query:"optional"`
}
```

Source: `lib/query/field_tag.go`, `lib/routing/op.go` (required enforcement),
`mod/nat/src/op_punch.go`

## Module Client

Every module with query ops has a `mod/<name>/client/` package mirroring the
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
| Module running on same node | Dependency interface, e.g. `mod.Dir.ResolveIdentity(name)` |
| Operation on a different node | Client with target, e.g. `natclient.New(target, astrald.Default())` |
| External app with no node access | Default client routed through apphost |

Source: `mod/nat/src/op_punch.go`, `mod/nat/src/op_node_consume_hole.go`
