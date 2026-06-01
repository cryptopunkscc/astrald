# Object Patterns

Use these patterns when defining typed payloads, storing objects, or integrating
with `objects` module extension points.

## Object Definition

Define a typed payload with:

* An `ObjectType` method.
* `WriteTo` and `ReadFrom` methods backed by `astral.Objectify`.
* Registration in `init`.

```go
type MyMsg struct {
    Token astral.Nonce
    Name  astral.String8
}

func (MyMsg) ObjectType() string { return "mod.mymodule.my_msg" }

func (msg MyMsg) WriteTo(w io.Writer) (int64, error) {
    return astral.Objectify(&msg).WriteTo(w)
}

func (msg *MyMsg) ReadFrom(r io.Reader) (int64, error) {
    return astral.Objectify(msg).ReadFrom(r)
}

func init() { _ = astral.Add(&MyMsg{}) }
```

Source: `mod/apphost/bind_msg.go`

Add JSON support only when needed:

```go
func (msg MyMsg) MarshalJSON() ([]byte, error)  { return astral.Objectify(&msg).MarshalJSON() }
func (msg *MyMsg) UnmarshalJSON(b []byte) error { return astral.Objectify(msg).UnmarshalJSON(b) }
```

Field type reference: `.ai/knowledge/concepts/wire.md`.

## Receiver

Accept only expected object types. Return `nil` for objects this receiver does
not handle.

```go
func (mod *Module) ReceiveObject(drop objects.Drop) error {
    switch obj := drop.Object().(type) {
    case *MyExpectedType:
        mod.handleIncoming(drop.SenderID(), obj)
        return drop.Accept(false)
    }
    return nil
}
```

`drop.Accept(save bool)` acknowledges the object.

* `save=true` stores it in `WriteDefault` at most once across all receivers.
* Omitting `Accept` silently ignores the object for this receiver.
* Other receivers still run.

Source: `mod/nat/src/object_receiver.go`, `mod/nodes/src/object_receiver.go`

## Describer

Check the zone before reading module state. Return descriptors through a channel.

```go
func (mod *Module) DescribeObject(ctx *astral.Context, objectID *astral.ObjectID) (<-chan *objects.Descriptor, error) {
    if !ctx.Zone().Is(astral.ZoneDevice) {
        return nil, astral.ErrZoneExcluded
    }

    rows, err := mod.db.FindByObjectID(objectID)
    if err != nil {
        return nil, err
    }

    results := make(chan *objects.Descriptor, len(rows))
    defer close(results)
    for _, row := range rows {
        results <- &objects.Descriptor{
            SourceID: mod.node.Identity(),
            ObjectID: objectID,
            Data:     &MyPayload{Path: astral.String16(row.Path)},
        }
    }
    return results, nil
}
```

Channel rule:

* Use a buffered channel and close it synchronously when results are known
  immediately.
* Use an unbuffered channel and close it from a goroutine when results arrive
  asynchronously.

Source: `mod/fs/src/object_describer.go`

## Searcher

Check the zone and required tags before starting the result stream.

```go
func (mod *Module) SearchObject(ctx *astral.Context, query objects.SearchQuery) (<-chan *objects.SearchResult, error) {
    if !ctx.Zone().Is(astral.ZoneDevice) {
        return nil, astral.ErrZoneExcluded
    }
    if err := query.RequiredTagsIn("path"); err != nil {
        return nil, err
    }

    results := make(chan *objects.SearchResult)
    go func() {
        defer close(results)
        rows, err := mod.db.SearchByPath(strings.ToLower(string(query.Query)))
        if err != nil {
            mod.log.Error("search: db: %v", err)
            return
        }
        for _, row := range rows {
            results <- &objects.SearchResult{
                SourceID: mod.node.Identity(),
                ObjectID: row.DataID,
            }
        }
    }()
    return results, nil
}
```

Always use an unbuffered channel with a goroutine for search results.

Source: `mod/fs/src/object_searcher.go`

## Finder

Return identities that can provide an object. Use the same buffered/goroutine
channel rule as `Describer`.

```go
func (mod *Module) FindObject(ctx *astral.Context, objectID *astral.ObjectID) (<-chan *astral.Identity, error) {
    providers, err := mod.db.providersFor(objectID)
    if err != nil {
        return nil, err
    }
    out := make(chan *astral.Identity, len(providers))
    defer close(out)
    for _, p := range providers {
        out <- p
    }
    return out, nil
}
```

## Holder

Fail closed: protect the object when the lookup itself fails. `objects.purge`
asks every registered holder before deleting; returning `true` keeps the object.

```go
var _ objects.Holder = &Module{}

func (mod *Module) HoldObject(objectID *astral.ObjectID) bool {
    held, err := mod.db.referencesObject(objectID)
    if err != nil {
        mod.log.Error("object hold lookup failed: %v", err)
        return true
    }
    return held
}
```

Source: `mod/auth/src/object_holder.go`, `mod/crypto/src/object_holder.go`, `mod/user/src/object_holder.go`, `mod/apphost/src/object_holder.go`

`objects.delete` bypasses holders; only `objects.purge` consults them.
