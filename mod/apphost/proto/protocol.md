# apphost protocol

apphost is a protocol that lets apps interact with the node.

## Overview

The client sends a command name (encoded as an 8-bit length-encoded string) followed by its arguments.

List of methods:

| name     | desc                              |
|----------|-----------------------------------|
| register | register a port on the local node |
| query    | send a query to a node by id      |
| resolve  | resolve node id from name         |
| nodeInfo | get info about a node             |

## Commands

### register

Arguments

| type   | name   | desc                                              |
|--------|--------|---------------------------------------------------|
| []byte | port   | port to register (8-bit LE string)                |
| []byte | target | where to forward connections to (8-bit LE string) |

`target` is in format `<proto>:<address>` where proto can be unix or tcp. 

Return values

| type | name  | desc       |
|------|-------|------------|
| byte | error | error code |

Error codes

| code | desc                |
|------|---------------------|
| 0x00 | no error            |
| 0x02 | registration failed |

### query

Arguments

| type     | name     | desc                   |
|----------|----------|------------------------|
| [33]byte | identity | remote node's identity |
| []byte   | query    | 8-bit LE query string  |

Return values

| type | name  | desc       |
|------|-------|------------|
| byte | error | error code |

Error codes

| code | desc           |
|------|----------------|
| 0x00 | no error       |
| 0x01 | query rejected |

If there was no error, the protocol ends and the connection is replaced with
the query connection.

### resolve

Arguments

| type   | name   | desc                              |
|--------|--------|-----------------------------------|
| []byte | port   | name to resolve (8-bit LE string) |

Return values

| type     | name     | desc              |
|----------|----------|-------------------|
| [33]byte | identity | resolved identity |
| byte     | error    | error code        |

### nodeInfo

Arguments

| type     | name     | desc                   |
|----------|----------|------------------------|
| [33]byte | identity | node's identity        |

Return values

| type     | name     | desc                          |
|----------|----------|-------------------------------|
| [33]byte | identity | node's identity               |
| []byte   | name     | node's name (8-bit LE string) |

