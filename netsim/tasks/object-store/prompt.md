You're running an astral node here that you control as its user. Store a short,
distinctive piece of text as an astral object on `__TARGET__`, and note the object
id you get back. Then read it back from `__TARGET__` and check the bytes match what
you stored.

When you're done, add the object id, the exact text you stored, and what you read
back to `~/info.json` (as `object_id`, `object_payload`, `object_readback`), leaving
any existing entries in the file in place.
