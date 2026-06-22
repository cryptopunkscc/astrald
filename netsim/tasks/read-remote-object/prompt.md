You're running an astral node here that you control as its user. There's an astral
object stored on `__PEER__` — its id is in `~/info.json` (the `object_id` value).
Read that object from `__PEER__` and check what it contains.

When you're done, add what you read to `~/info.json` (as `object_remote`), leaving
the existing entries in the file in place.
