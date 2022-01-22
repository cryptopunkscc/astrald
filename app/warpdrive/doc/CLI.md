# Warp Drive CLI v1.0.0-dev

Command line service for warpdrive, combined with [ANC](../../../cmd/anc/README.md) (astral netcat) can be considered as
referential UI client for development and testing purpose. Under the hood, it translates text commands directly into
warpdrive UI API calls and returns received results as formatted console output. Depending on needs is recommended to
use more than one client instance connected at the same time, this can be useful for tracking subscription updates and
triggering updates at the same time.

### Running console client

ANC is a commandline astral client, that allows querying services registered under specified port, on local or remote
device:

```shell
$ anc query <identity> <port>
```

The warpdrive CLI allows only for localhost connections, so you have to omit the identity as redundant, and leave only
the correct warpdrive port.

For example, to connect to warpdrive CLI, you can run ANC directly from source file [main.go](../../../cmd/anc/main.go):

```shell
$ go run ./cmd/anc/main.go query wd
connected.
```

# Features

Complete list of warpdrive CLI features represented as console output.

## `sender`

### `peers` aka `recipients`

```shell
warp> peers
<peer_id> <peer_alias>
...
ok
```

example:

```shell
warp> peers
02f978d6bd70d0005f5148ce5h311609c994219164126488f11573440b2c6a40eb localhost  
024d47047667312be7cd0a140f3323b716030f5fc9d37ae774eb96527a76fa55f9 remote
ok
```

### `send`

```shell
warp> send <file_uri> <peer_id>
<offer_id>
ok
```

example:

```shell
warp> send ./myfiles/archive1.zip 02f978d6bd70d0005f5148ce5h311609c994219164126488f11573440b2c6a40eb
3a076839-1b17-41a9-50b9-72beee2d08db
ok
```

### `sent`

```shell
warp> sent
<offer_id> <status>
  -  <file_uri>
  -  ...
...
ok
```

### `events`

```shell
warp> events sent
<offer_id> <sent|rejected|accepted|uploaded>
...
```

example:

```shell
warp> events sent
3a076839-1b17-41a9-50b9-72beee2d08db sent
3a076839-1b17-41a9-50b9-72beee2d08db rejected
...
```

## `recipient`

### `received`

```shell
warp> received
<offer_id> <received|rejected|accepted|downloaded>
  -  <file_uri>
  -  ...
...
ok
```

### `accept`

```shell
warp> accept <offer_id>
accepted <offer_id>
ok
```

example:

```shell
warp> accept <offer_id>
accepted
ok
```

### `reject`

```shell
warp> reject <offer_id>
rejected
ok
```

### `update`

```shell
warp> update <peer_id> mod <trust|block|"">
updated
ok
```

### `events`

```shell
warp> events received
<offer_id> <received|rejected|accepted|downloaded>
...
```