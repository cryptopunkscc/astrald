# read-remote-object

node1's agent reads a peer's object (id from `~/object.json`) over astral as the User and records the bytes to `~/read.json`. verify re-reads the peer's object via `<peer>:objects.load` and asserts the bytes equal node1's stored `payload.txt`. Produces the remote read in `read-remote-peer.story`.
