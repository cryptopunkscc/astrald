# object-store

Drives the Qwen operator on node1 to **store an astral object and read it back**,
following the astral-agent skill, in one of two modes (`--target`):

- `self` (default): store in node1's **own local repo** — the basic local object
  operations (`objects.store` then `objects.load`).
- `peer`: store **on the sibling node2** (`<node2>:objects.store`) — write to a peer.

`verify.py` independently re-loads the object from the **holder's** local repo
(`objects.load -repo local`, ungated) and asserts the bytes match. Stories:

- `object-store.story` (self) → `two-nodes-data` (object on node1) — feeds
  `read-remote-object`.
- `object-store-peer.story` (peer) → `two-nodes-data-peer` (object on node2).
