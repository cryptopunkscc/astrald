# apphost protocol

Version 1.0

## Overview

**THIS PROTOCOL DESCRIPTION IS OBSOLETE**

The `apphost` protocol is a client-server multiple request-response protocol
used by apps (guests) to access the astral network via the local node (host).

The guest connects to the host via any of the supported [IPC](#ipc) methods.
If the host accepts, the session is established, and it ends whenever the
connection is closed. Over the duration of the sesssion the guest can at any
point send a request to the host, after which the guest must wait for the host's
response. If the request did not result in session termination or state change,
the guest can use the same connection to send further requests.

Requests are encoded as String8 followed by their argumets.

## Table of contents

* [Methods](#methods)
  * [token](#token)
  * [register](#register)
  * [query](#query)
* [Messages](#messages)
  * [queryInfo](#queryinfo)
* [Types](#types)
  * [Basic types](#basic-types)
  * [Identity](#identity)
* [IPC](#ipc)
  * [TCP](#tcp)
  * [Unix sockets](#unix)

## Methods

### token

`(token String8) -> (code Uint8, guestID Identity, hostID Identity)`

The `token` request authenticates the session with an auth token.

#### Arguments

| name  | type    | desc           |
|:------|:--------|:---------------|
| token | String8 | the auth token |

#### Return values

| name     | type      | desc                                           |
|:---------|:----------|:-----------------------------------------------|
| code     | Uint8     | return code                                    |
| guestID  | Identity  | only if code is 0, the identity of the guest   | 
| hostID   | Identity  | only if code is 0, the identity of the host    | 

#### Return codes

| code | desc        |
|:-----|:------------|
| 0x00 | success     |
| 0x01 | auth failed |

### register

`(endpoint String8) -> (code Uint8, token String8)`

Register sets the query handler address for the guest's identity. The host
will forward all queries directed to the guest to this endpoint.

If the registration succeeds, return code 0 will be returned and the connection
will go into keep-alive mode with no data sent. Once either side closes the
conenction the registration expires. If the registration fails, the session
continues.

After the query handler accepts the callback connection and the endpoint,
it must read the `queryInfo` message. The handler has two options on how to
handle the query - skip it by closing the connection or respond to it by
sending an Uint8 response code. If the handler decides to skip the query, it
will be passed to the next registered handler (if any). If the handler responds
with code `0`, the connection becomes the transport for the accepted query.
If the response code is non-zero, the connection is closed.

#### Arguments

| name     | type    | desc                               |
|:---------|:--------|:-----------------------------------|
| endpoint | String8 | query handler [IPC](#ipc) endpoint |
| flags    | Uint8   | always 0. reserved.                |

`endpoint` is in format `<proto>:<address>` where proto can be Unix or TCP.

#### Return values

| name  | type    | desc                                           |
|:------|:--------|:-----------------------------------------------|
| code  | Uint8   | return code                                    |
| token | String8 | auth token for callbacks, sent only on success |

#### Return codes

| code | desc               |
|:-----|:-------------------|
| 0x00 | success            |
| 0x01 | unauthorized       |
| 0x02 | already registered |

### query

`(target Identity, query String16) -> (code Uint8, ...)`

The `query` method routes a query through the node. If the query is rejected,
the session continues. If the query is accepted, the session ends and the
connection becomes the transport for the accepted query.

#### Arguments

| name     | type     | desc             |
|:---------|:---------|:-----------------|
| identity | Identity | target identity  |
| query    | String16 | the query string |

#### Return values

| name | type | desc        |
|:-----|:-----|:------------|
| code | byte | return code |

#### Return codes

| code      | desc                  |
|:----------|:----------------------|
| 0x00      | accepted              |
| 0x01      | query rejected        |
| 0x02-0xff | query-specific errors |

If there was no error, the protocol ends and the connection is replaced with
the query connection.

## Messages

### queryInfo

The query message is the first message sent to the query handler registered
with the `register` method.

#### Fields

| name    | type      | desc                                           |
|:--------|:----------|:-----------------------------------------------|
| token   | String8   | auth token obtained via the `register` call    |
| caller  | Identity  | identity of the caller                         |
| query   | String16  | the query string                               |

## Types

Numeric types use big endian encoding.

### Basic types

The basic integer types are Uint8, Uint16, Uint32, Uint64, Int8, Int16, Int32,
Int64.

String types (String8, String16, String32, String64) represent a length encoded
string using 8/16/32/64-bit unsigned integers.

### Identity

Identity is a fixed-length buffer of 33 bytes and contains the public key of the
identity.

## IPC

Guests can use various IPC methods to establish a connection to the host.
The endpoints have format of "method:address", for example "tcp:127.0.0.1:8080".
Supported methods are tcp and unix.

### tcp

Standard TCP address, for example "127.0.0.1:1234".
By default, apphost listens on "127.0.0.1:8625".

### unix

A unix socket, for example "/var/run/app.socket".
By default, apphost listens on "~/.apphost.sock".

### WebSocket

The HTTP server (default `tcp:0.0.0.0:8624`) also accepts WebSocket upgrades on
`/.ws`, intended for browser-based guests. The endpoint is loopback-only —
non-loopback requests are refused. There is no TLS; if you need WS from a
non-loopback origin, terminate TLS in a reverse proxy.

Mode is selected by `Sec-WebSocket-Protocol`:

| Subprotocol         | Frame   | Payload format                                     |
|---------------------|---------|----------------------------------------------------|
| `astral.binary.v1`  | binary  | the existing `String8 type + Bytes32 payload` byte stream, identical to TCP/unix |
| `astral.json.v1`    | text    | one JSON-encoded `astral.JSONAdapter` (`{"Type":"...","Object":...}`) per frame  |

If the client requests neither, the upgrade is closed with
`StatusPolicyViolation` (1008).

Origin policy: any `Origin` is accepted at the WebSocket layer. The endpoint
is already loopback-only at the network layer (non-loopback peers get HTTP
403), so a remote page cannot reach it regardless of Origin. For queries
arriving over the WebSocket, the browser `Origin` is attached to the
`InFlightQuery` under the `origin-web` key in `Extra`, leaving per-origin
authorization to the individual ops.

Authentication uses the same in-protocol `AuthTokenMsg` as the TCP/unix path.
The HTTP `Authorization: Bearer` header used by the rest of the HTTP bridge
is not consulted for `/.ws`.

JSON-mode auto-injection: when a JSON-mode guest sends `RouteQueryMsg`, the
server appends `out=json&in=json` to the query string if those params are
absent, so the responder's channel produces JSON-Lines for the post-accept
stream. Apps can override either by setting them explicitly in the query.
Channel-aware responders that emit `astral.Bytes32` get base64 in JSON
automatically; responders that bypass the channel and write raw bytes are
not safe for JSON mode and should be queried over `astral.binary.v1`
instead.

Outbound queries: one query per WS — closing the WS cancels the in-flight
query. The native `apphost.register_handler` op is not usable over WS
because it expects an IPC dial endpoint; WS clients use the registration
flow below instead.

#### Handling inbound queries

A WS client can register as a service handler for an identity it owns. The
host then pushes a notification per inbound query, and the client opens a
fresh per-query WS to respond.

**Registration WS** (long-lived, one per service identity):

1. Standard handshake (`HostInfoMsg` → optional `AuthTokenMsg` → `AuthSuccessMsg`).
2. Client → `RegisterServiceMsg{Identity}`.
3. Host → `Ack`. Authorization mirrors `RouteQueryMsg`: caller must equal
   Identity or hold a `SudoAction` for it.
4. For each inbound query targeting Identity, host pushes
   `IncomingQueryMsg{QueryID, Caller, Target, Query}`.
5. For each notification the client must either:
   - open a per-query WS within 5 seconds (see below), or
   - send `RejectIncomingMsg{QueryID, Code}` on the registration WS, or
   - ignore — the caller sees route-not-found after the timeout.
6. Closing the registration WS unregisters the handler.

**Per-query WS** (short-lived, one per accepted inbound query):

1. Standard handshake (`HostInfoMsg`). No auth needed — the unguessable
   `QueryID` is the pairing token.
2. Client → `AttachQueryMsg{QueryID}`.
3. Host → `Ack` (or `ErrorMsg{Code: route_not_found}` if the QueryID is
   unknown or already expired).
4. The connection becomes the bidirectional bytestream for that one query,
   with the client acting as the responder. Format is whatever the
   subprotocol selected; in `astral.json.v1` mode it's JSON-Lines of
   `astral.JSONAdapter`.
5. Closing the WS ends the query.