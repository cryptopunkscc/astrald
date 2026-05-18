# Project Overview

## Summary

Astrald is a proof-of-concept P2P network (Go 1.24+) providing authenticated encrypted connections over multiple transports. Core model: Identities (secp256k1 keys) expose named Services; Queries open bidirectional Sessions; Nodes establish encrypted Links; Objects are immutable content-addressed.

## Project Structure

```text
astral/     core types: Identity, Object, ObjectID, Query, Router, Zone
brontide/   Noise XK protocol
core/       Node, PriorityRouter, module manager
streams/    stream utilities
sig/        thread-safe collections: Map, Set, Queue, Ring, Pool
lib/        client libraries: apphost, astrald, query, routers, ipc
mod/        pluggable modules
cmd/        binaries
```

## Config

- Linux: `$HOME/.config/astrald/`
- macOS: `~/Library/Application Support/astrald/`

Main config: `node.yaml`. Per-module config: `<name>.yaml`.
