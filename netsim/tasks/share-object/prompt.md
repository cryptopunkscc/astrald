On this machine an `astrald` node is running and you control it as its User (a
User-bound apphost token is at `~/.netsim/user.token`). Acting as that User,
store a short, distinctive text payload as an astral object via the objects
protocol, following your **astral-agent** skill, and note the Object ID it
returns.

Then write that Object ID to `~/.netsim/object.id` and the exact payload you
stored to `~/.netsim/object.payload`. The skill won't mention these files — they
are how the run is checked. Success means an Object ID is returned and both files
are written.
