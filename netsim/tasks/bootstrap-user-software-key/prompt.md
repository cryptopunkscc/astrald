On this machine there is an `astrald` node running. It has its own node identity
but no User. Make it a User-controlled node under a fresh software User, following
your **astral-agent** skill's node-setup playbook.

Then write the User's id and a User-bound apphost token to `$HOME/info.json` as a
JSON object with keys `user_id` and `user_token`. The skill won't mention this —
it's how the run is checked.
