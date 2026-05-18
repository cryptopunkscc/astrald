# Relay

## Definition

* A `Relay` is a third-party node that forwards queries between two peers that
  cannot connect directly.
* A relay operates at the query layer, not the transport layer.
* A relay is distinct from the `gw` transport, which works at the byte-stream
  layer.

## Service Flow

* A relay node exposes `nodes.node_open_relay` (`MethodNodeOpenRelay`).
* Clients open that service and stream `QueryContainer` values over the
  resulting channel.
* The relay server matches each container against a waiting peer session and
  pipes the two sides together.

## Wire Objects

* `QueryContainer` is `{TargetID, CallerID, Query}`.
* The initiating side sends `QueryContainer` to name the target and declare the
  originating caller.

## Relay Channel Cache

* `relayChannel` is the cached `Router` for a given relay identity.
* `getRelay` opens it lazily on first use.
* Its `watch` goroutine evicts it after 30 min idle.
* A broken send evicts the channel immediately.

## Routing

* Query routing selects the relay path when `ctx.Identity() != q.Caller`, even
  when the target peer is already linked.
* This condition checks caller identity, not connectivity.
* `ExtraRelayVia` is the key into `q.Extra`; its value is
  `[]*astral.Identity`.
* Relays in `ExtraRelayVia` are tried in order when direct and NAT link
  strategies both fail.
* The target identity is never used as its own relay.

## Authorization

* `RelayForAction` is the typed auth action.
* `ForID *Identity` names the caller whose relay queue the actor requests to
  join.
* `ActionRelayFor = "mod.nodes.relay_for_action"`.
* `OpNodeOpenRelay` checks it when `container.CallerID != q.Caller()`.
* `AuthorizeRelayFor` grants when `actor == ForID`.
* `RelayForAction` implements `Constrainable`.
* `ApplyConstraints` returns true when the constraint bundle is nil or empty.
