# apphost protocol

Version 1.0

## Overview

The `apphost` protocol is a client-server multiple request-response protocol
used by apps (guests) to access the astral network via the local node (host).

The guest connects to the host via any of the supported [IPC](#ipc) methods.
After the connection is established, the host sends a `version` message to the
guest. At this point the session is established, and it ends whenever the
connection is closed. Over the duration of the sesssion the guest can at any
point send a request to the host, after which the guest must wait for the host's
response. If the request did not result in session termination or state change,
the guest can use the same connection to send further requests.

## Table of contents

* [Methods](#methods)
  * [token](#token)
  * [register](#register) 
  * [query](#query)
* [Messages](#messages)
  * [version](#version)
  * [queryInfo](#queryinfo)
* [Types](#types)
  * [Basic types](#basic-types)
  * [Identity](#identity)
* [IPC](#ipc)
  * [TCP](#tcp)
  * [Unix sockets](#unix-sockets)

## Methods

### token

`(token String8) -> (code Uint8)`

The `token` request authenticates the session with an auth token.

#### Arguments

| name  | type    | desc           |
|-------|---------|----------------|
| token | String8 | the auth token |

#### Return values

| name     | type     | desc                                             |
|----------|----------|--------------------------------------------------|
| code     | Uint8    | return code                                      |
| identity | Identity | on success, the identity assigned to the session | 

#### Return codes

| code | desc        |
|------|-------------|
| 0x00 | success     |
| 0x01 | auth failed |

### register

`(endpoint String8) -> (code Uint8, token String8)`

Register sets the query handler address for the guest's identity. The host
will forward all queries directed to the guest to this port.

If the registration succeeds, return code 0 will be returned and the connection
will go into keep-alive mode with no data sent. Once either side closes the
conenction the registration expires.

After the query handler accepts the callback connection and the endpoint,
it must read the `queryInfo` message and respond with a single Uint8 value
representing the response code. If the code sent is 0, the connection becomes
the transport for the accepted query. In any other case, the connection is
terminated.

#### Arguments

| name     | type    | desc                   |
|----------|---------|------------------------|
| endpoint | String8 | query handler endpoint |
| flags    | Uint8   | always 0. reserved.    |

`endpoint` is in format `<proto>:<address>` where proto can be Unix or TCP.

#### Return values

| name  | type    | desc                                           |
|-------|---------|------------------------------------------------|
| code  | Uint8   | return code                                    |
| token | String8 | auth token for callbacks, sent only on success |

#### Return codes

| code | desc               |
|------|--------------------|
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
|----------|----------|------------------|
| identity | Identity | target identity  |
| query    | String16 | the query string |

#### Return values

| name | type | desc        |
|------|------|-------------|
| code | byte | return code |

#### Return codes

| code      | desc                  |
|-----------|-----------------------|
| 0x00      | accepted              |
| 0x01      | query rejected        |
| 0x02-0xff | query-specific errors |

If there was no error, the protocol ends and the connection is replaced with
the query connection.

## Messages

### version

Version is a single Uint8 value. Currently, it's always 1.

### queryInfo

The query message is the first message sent to the query handler registered
with the `register` method.

#### Fields

| name    | type      | desc                                          |
|:--------|:----------|:----------------------------------------------|
| token   | String8   | auth token obtained via the `register` call   |
| caller  | Identity  | identity of the caller                        |
| query   | String16  | the query string                              |Info

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
Currently supported methods are TCP and Unix sockets.

### TCP

TBD

### Unix sockets

TBD