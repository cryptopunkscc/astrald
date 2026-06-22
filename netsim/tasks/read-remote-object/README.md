# read-remote-object

Has **node1's agent read an astral object that lives on the peer** (node2), over
astral. The object id is in node1's `~/info.json` (`object_id`, written by
`object-store --target node2`); the agent reads it from the peer **as the User**
(addressing the peer by its alias from `adopt-node`) and records what it read.
`verify.py` independently re-reads the peer's object as the User and asserts the
bytes match.

Used by `read-remote-peer.story` (which first stores the object on node2, then runs
this read). Note: an *anonymous* read of a peer's object does **not** route (the
network zone is stripped); the read must come from an authenticated identity, which
is why it's driven by the User on node1.
