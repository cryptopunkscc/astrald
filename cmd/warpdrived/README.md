# warp drive daemon

Command line runner for [lib:warpdrive](../../lib/warpdrived).

## How to run

Make sure the astral node is running. If not start it:

```shell
go run ./cmd/astrald
```

Run warp drive service:

```shell
go run ./cmd/warpdrived
```

## ANC CLI

Warp Drive daemon serves commandline interfaces on `wd` port. 
You can use astral netcat tool to connect interactive cli session:

```shell
go run ./cmd/and quert [node id] wd
```
