On this machine an `astrald` node is running and you control it as its User (a
User-bound apphost token is at `~/.netsim/user.token`). Your swarm has one other
node — a sibling. Acting as that User, store a short, distinctive text payload as
an astral object **on that other node** — address it explicitly as the query
target — via the objects protocol, following your **astral-agent** skill, and note
the Object ID it returns. Then read the object back **from that other node** by its
Object ID and confirm the bytes match what you stored.

Then write the Object ID to `~/.netsim/object.id`, the exact payload you stored to
`~/.netsim/object.payload`, the bytes you read back to `~/.netsim/object.readback`,
and the node id you stored it on to `~/.netsim/object.target`. The skill won't
mention these files — they are how the run is checked. Success means the object is
stored on the other node, read back with matching bytes, and all four files are
written.
