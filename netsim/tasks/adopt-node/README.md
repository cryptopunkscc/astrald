# adopt-node

Drives the Qwen operator on node1 to **adopt node2** into the User's swarm
(`user.adopt`), following the astral-agent skill's node-adoption playbook.
`verify.py` independently confirms both nodes hold a contract under the same User,
a mutual link, and a symmetric roster (each lists the other as a Linked sibling).
Produces stage `astrald-swarm` (from `astrald-single-node`).
