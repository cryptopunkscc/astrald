On this machine there is an `astrald` node running. It has its own node identity
but no User. Make it a User-controlled node under a fresh software User, following
your **astral-agent** skill's node-setup playbook.

Then write the User's id to `~/.netsim/user.id` and a User-bound apphost token to
`~/.netsim/user.token`. The skill won't mention this — it's how the run is checked.
