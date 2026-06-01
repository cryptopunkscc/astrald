# Query

A `Query` requests a bidirectional `Session` with a named service on a target
`Identity`. It is the base communication operation.

## Types

* `Query{Nonce, Caller, Target, QueryString}` is the wire object.
* `InFlightQuery` wraps a `Query` with `Extra sig.Map[string, any]`.
* `OriginLocal` and `OriginNetwork` are values for `Extra["origin"]`.
* `IsLocal()` and `IsNetwork()` test that key.

## Routing

* `Router.RouteQuery(ctx, *InFlightQuery, w) (io.WriteCloser, error)`.
* Node runs registered routers in priority order.
* Outcomes: accept opens the session, reject stops routing, not found tries
  the next router.
* `ErrRouteNotFound` is now an empty struct; matchers use `errors.Is`.

## Preprocessors

* Preprocessors run before routing, in registration order.
* They may attach metadata, add `relay_via` hints, or block the query.
* `core` discovers them via `injectLoaded`. apphost uses one to attach
  AppContracts.

## Gateway

Gateway relays for nodes unreachable directly (NAT, firewall). The
application sees a normal session.
