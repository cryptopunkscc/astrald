# share-object

Drives the Qwen operator on node1 to **store an astral object on its swarm sibling
node2** (`<node2>:objects.store`) and read it back. `verify.py` independently
confirms node2 physically holds the object in its local repo. Produces stage
`astrald-shared` (from `astrald-swarm`).
