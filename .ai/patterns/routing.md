# Routing Patterns

Use when implementing routers, forwarding queries, or enforcing zone scope.

## Zone Handling

Gate all network operations with `ZoneNetwork`.

- A hop may narrow zones.
- A hop must not expand zones.

```go
if !ctx.Zone().Is(astral.ZoneNetwork) {
    return query.RouteNotFound()
}

ctx = ctx.ExcludeZone(astral.ZoneNetwork)
ctx = ctx.WithZone(astral.ZoneDevice)
```

Modules that require network access include it in `Run`.

```go
func (mod *Module) Run(ctx *astral.Context) error {
    mod.ctx = ctx.IncludeZone(astral.ZoneNetwork)
    return nil
}
```

Source: `mod/apphost/src/guest.go`, `mod/nodes/src/module.go`

## RouteQuery Return Values

`astral.Router.RouteQuery(ctx, *astral.InFlightQuery, w io.WriteCloser)` returns
`(io.WriteCloser, error)`. Return the first matching result. Never fall through
with `nil, nil`.

```go
func (r *MyRouter) RouteQuery(ctx *astral.Context, q *astral.InFlightQuery, w io.WriteCloser) (io.WriteCloser, error) {
    if !r.matches(q) {
        return query.RouteNotFound()
    }
    if !r.authorized(q) {
        return query.Reject()
    }
    return query.Accept(q, w, func(conn astral.Conn) {
        defer conn.Close()
    })
}
```

| Situation | Return |
|---|---|
| Not our query | `query.RouteNotFound()` |
| Explicit refusal | `query.Reject()` (or `query.RejectWithCode(code)`) |
| Accepted | `query.Accept(q, w, handler)` |
| Never | `nil, nil` |

`query.RouteNotFound` and `query.Reject` take no arguments. `ErrRouteNotFound`
carries no router reference; routers identify themselves through their own
`String()` or `Name`.

Source: `lib/query/route.go`, `astral/err_route_not_found.go`
