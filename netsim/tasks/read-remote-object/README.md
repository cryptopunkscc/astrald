# read-remote-object

Confirms a peer (node2) can read **node1's** object **over astral** — the object
that `object-store` stored in node1's local repo. Host-driven: node2 has no Qwen
operator, so `verify.py` issues the read (resolve node1's identity + the stored
`object_id`, then node2 runs `<node1>:objects.load` and asserts the exact bytes,
with transparent/`objects.find` as fallback diagnostics). No agent comprehension
axis — this is a pure implementation-axis probe. Produces stage `astrald-read`
(from `astrald-stored`).

Note: the peer-reads-node1 direction **failed before astrald #348** (the roster
sync); this task re-probes it on current master. It may now pass (node2 knows
node1, and `op_load` is ungated) or surface the gap — either is a valid finding.
