# expel-node

Drives the Qwen operator on node1 (the swarm's User) to **permanently expel node2**
from the swarm (`user.expel`), following the astral-agent skill's knowledge of the
user protocol. Expelling bans the node's identity: it is dropped from the swarm
roster and its links are torn down, though its underlying membership contract is
**not** revoked — the ban is enforced by a membership filter, not contract removal.

`verify.py` independently confirms the post-ban state from both ends: node2 is
recorded in node1's `user.list_expelled`, node2 no longer appears in node1's
`user.swarm_status` (the roster shrinks — `OpSwarmStatus` lists `ActiveNodes`,
which filters the `expelledSet`), and the node1↔node2 link is gone on both ends.
Produces stage `two-nodes-expel` (from `two-nodes`).
