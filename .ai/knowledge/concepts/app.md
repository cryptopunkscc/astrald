# App

An `App` is an external process that uses a Host `Node` through the `apphost` module. The Host owns routing, transport, encryption, and zone enforcement; the App has no direct network presence and speaks the apphost protocol over an IPC bridge or a loopback WebSocket.

## Bridges

The same wire protocol is offered on two transports:

* IPC: TCP / unix / memu listeners from `config.Listen`.
* WebSocket: loopback-only upgrade at `/.ws` on the HTTP bind, with `astral.binary.v1` (binary frames) or `astral.json.v1` (text frames) as subprotocol.

A third bridge, the bearer-auth HTTP gateway, is not the apphost protocol; it exposes objects under `/.objects/<id>` and translates arbitrary URLs into single queries.

## Identity model

* `Identity` — the network identity the App acts as.
* `GuestID` — the identity bound to a live connection. Zero (anonymous) until authentication.
* `AccessToken` — `{Identity, Token, ExpiresAt}`. Time-bounded host-issued credential that binds an `Identity` to a connection on AuthTokenMsg.

Authentication:

* No `AuthTokenMsg` -> `GuestID` stays zero. Anonymous guests can route only when `allow_anonymous` is true and always lose `ZoneNetwork`.
* Valid `AuthTokenMsg` -> `GuestID = token.Identity`. Multiple Apps may connect with different GuestIDs.
* Static tokens come from `config.Tokens` (resolved by `dir`, 100-year expiry). Dynamic tokens come from `apphost.create_token` (32-char random, default 1-year expiry) or `apphost.register`.

A guest may act as another identity only when `auth.Authorize(SudoAction{Actor:GuestID, AsID:target})` grants it. This gate covers `Caller` override on outbound queries and identity override on both IPC and WS handler registration.

## Capabilities

* Send queries (`RouteQueryMsg`): the Host enforces zones, attaches relay contracts, and bridges the session back as a byte stream after `QueryAcceptedMsg`.
* Register an IPC handler: the Host opens a fresh IPC dial to `Endpoint` for each inbound query and pushes `HandleQueryMsg`.
* Register a WS service handler: the Host pushes `IncomingQueryMsg` on the registration WS; the App opens a per-query WS and sends `AttachQueryMsg` to accept, or `RejectIncomingMsg` on the registration WS to refuse. Default attach timeout is 5 s.
* Cancel in-flight queries by nonce.
* Hold and unhold object IDs to block purge.

Handlers disappear on disconnect (the IPC guest-conn variant, the WS service handler) or when `bind` releases their token.

## App record

`App` (`mod.apphost.app`) is the installed-app row: `{AppID, HostID, InstalledAt}`. Stored in `apphost__local_apps` by `apphost.install_app` with `OnConflict{DoNothing}`.

## AppContract

`AppContract` is an `auth.SignedContract` with:

* `Issuer` = AppID.
* `Subject` = HostID (the node).
* `Permits` granting `RelayForAction`.
* `ExpiresAt` from the requested duration (default 1 year, 10 years for `register`).

Uses:

* Relay authorization: the query preprocessor attaches the contract to outbound queries whose `Caller` matches the issuer.
* Relay hints: the preprocessor adds every non-local subject of a contract issued by `Target` as a relay hop.
* Identity proof in the local swarm: `User.PushToLocalSwarm` republishes signed contracts after `sign_app_contract` and `install_app`.

Three ops produce contracts:

* `apphost.new_app_contract` returns an unsigned `Contract`.
* `apphost.sign_app_contract` signs, indexes, stores, and pushes a caller-supplied `Contract`.
* `apphost.install_app` does the full path (build + sign + index + store + `CreateLocalApp` + push). Network-origin queries are rejected.

## Holds

A Hold is a row in `apphost__object_holds` keyed by `(AppID, ObjectID)`. While at least one active hold exists for an object, `objects.purge` skips it (the apphost `Module` exposes `HoldObject(objectID) bool` as an `objects.Holder` and is auto-registered by the objects module).

* `apphost.hold_object` inserts a hold for the caller; `Duration` is optional (`nil` -> no expiry, otherwise `hold_until = now + Duration`). `OnConflict{DoNothing}` makes repeated holds idempotent.
* `apphost.unhold_object` deletes only the caller's row for the given object.
* `apphost.list_held_objects` streams the caller's active holds (`hold_until IS NULL OR hold_until > now`) and ends with `EOS`.
* Network-origin queries are rejected; caller must be non-zero. Multiple apps may hold the same object.
