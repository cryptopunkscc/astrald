# object-store

Drives the Qwen operator on node1 to **store an astral object and read it back**,
following the astral-agent skill, on a chosen **target node** — `--target` is an
astral query target:

- `localnode` (default): store on the local node (node1's own repo) — basic local
  object operations (`objects.store` / `objects.load`).
- `node2` (a node alias registered by `adopt-node`): store on that node
  (`node2:objects.store`) — write to a peer.

`verify.py` independently re-loads the object from the **holder's** local repo
(`objects.load -repo local`, ungated) and asserts the bytes match (`localnode`/
`node1` → node1, `node2` → node2). Stories:

- `object-store.story` (`localnode`) → `two-nodes-data` (object on node1) — feeds
  `read-remote-object`.
- `object-store-peer.story` (`--target node2`) → `two-nodes-data-peer` (object on node2).
