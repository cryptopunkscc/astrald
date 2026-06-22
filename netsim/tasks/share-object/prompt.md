On this machine an `astrald` node is running and you control it as its User (a
User-bound apphost token is recorded in `$HOME/info.json` under `user_token`). Your
swarm has one other node — a sibling. Acting as that User, store a short,
distinctive text payload as an astral object **on that other node** — address it
explicitly as the query target — via the objects protocol, following your
**astral-agent** skill, and note the Object ID it returns. Then read the object
back **from that other node** by its Object ID and confirm the bytes match what you
stored.

Then add to `$HOME/info.json` (keep the existing `user_*` keys) these keys:
`object_id` (the Object ID), `object_payload` (the exact payload you stored),
`object_readback` (the bytes you read back), and `object_target` (the node id you
stored it on). The skill won't mention this — it's how the run is checked. Success
means the object is stored on the other node, read back with matching bytes, and
those keys are written.
