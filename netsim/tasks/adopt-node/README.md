# adopt-node

Drives the Qwen operator on node1 to **adopt node2** into the User's swarm
(`user.adopt`), following the astral-agent skill's node-adoption playbook.
`verify.py` independently confirms both nodes hold a contract under the same User,
a mutual link, and a symmetric roster (each lists the other as a Linked sibling).
Also registers `node1`/`node2` directory aliases (`dir.set_alias`) on both nodes so
later tasks can address nodes by name (e.g. `object-store --target node2`). Produces
stage `two-nodes` (from `one-node`).
