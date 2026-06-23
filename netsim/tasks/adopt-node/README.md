# adopt-node

Adopts node2 into node1's User swarm and registers `node1`/`node2` aliases on both nodes.

`verify.py` proves both nodes hold a contract from the same User, node2 links back to node1, the roster is symmetric (each lists the other as a Linked sibling), and `sibling_ids` includes node2.

Produces stage `two-nodes` (from `one-node`).
