# expel-node

The User (node1) permanently bans node2 from the swarm via `user.expel`. `verify.py` proves node2 is in `user.list_expelled`, gone from `user.swarm_status`, and the node1↔node2 link is torn down on both ends. From `two-nodes`; produces `two-nodes-expel`.
