# object-store

Drives the Qwen operator on node1 to **store an astral object in its own local
repo and read it back** — the basic local object operations (`objects.store` then
`objects.load`) — following the astral-agent skill. `verify.py` independently
re-loads the object from node1's local repo and asserts the bytes match. Produces
stage `two-nodes-data` (from `two-nodes`); the saved object is what
`read-remote-object` then fetches from a peer.
