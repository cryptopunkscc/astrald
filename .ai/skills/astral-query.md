# astral-query

Opens a query to a running node and pipes stdin/stdout through it. Use to call ops, inspect state, and observe typed object streams. For format theory, see `.ai/knowledge/concepts/wire.md`.

## Syntax
```sh
astral-query [caller@][target:]method [-key value]...
```

- `caller@` — optional
- `target:` — optional; falls back to `ASTRAL_DEFAULT_TARGET`
- `-key value` pairs passed as op arguments
- stdin → query channel; stdout ← response stream

## Format Args (`-out` / `-in`)

`-out` and `-in` are op arguments, not CLI flags. `astral-query` is a raw byte pipe; the op handler configures its channel:

```go
ch := channel.New(q.Accept(), channel.WithOutputFormat(args.Out))
```

Not every op supports `-out`/`-in` — check the op's arg struct in `op_*.go`.

## Op Discovery

**From source**: `mod/<module>/src/op_<name>.go` → `<module>.<name>`

**At runtime**:
```sh
astral-query shell.ops -out text
astral-query shell.ops -out text | grep nodes
```

## Useful Examples
```sh
astral-query nodes.links -out json
astral-query nodes.sessions -out json
astral-query objects.repositories -out json
astral-query objects.scan -repo main -out json
astral-query tree.get -path /mod/tcp/listen -out text
```

## Environment
- `ASTRAL_DEFAULT_TARGET` — default target identity
- `ASTRALD_APPHOST_TOKEN` — apphost auth token
- `ASTRAL_DEFAULT_INPUT_FORMAT` / `ASTRAL_DEFAULT_OUTPUT_FORMAT` — default `-in`/`-out`

## Source Paths
- `cmd/astral-query/main.go`
- `lib/ops/set.go`, `lib/ops/op.go`
