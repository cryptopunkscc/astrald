On this machine an `astrald` node is running and you control it as its User (a
User-bound apphost token is recorded in `$HOME/info.json` under `user_token`).
Acting as that User, store a short, distinctive text payload as an astral object
via the objects protocol, following your **astral-agent** skill, and note the
Object ID it returns. Then read that object back by its Object ID and confirm the
bytes match what you stored.

Then add to `$HOME/info.json` (keep the existing `user_*` keys) the keys
`object_id` (the Object ID), `object_payload` (the exact payload you stored), and
`object_readback` (the bytes you read back). The skill won't mention this — it's
how the run is checked. Success means the object is stored, read back with matching
bytes, and those keys are written.
