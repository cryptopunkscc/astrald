On this machine there is an `astrald` node running. It has its own node identity
but no User. You already control a software User whose BIP-39 mnemonic seed phrase
is:

  horse soldier imitate stool square buyer verb party enjoy result jazz rabbit trigger file benefit cloth term change

Make this node a User-controlled node under THAT existing User: derive the User's
`secp256k1` key from the mnemonic above (start from the mnemonic — do NOT generate
new entropy), then build, sign, and install the node contract, following your
**astral-agent** skill's node-setup playbook (software User) but substituting the
given mnemonic for the entropy-generation step.

Then write the User's id and a User-bound apphost token to `$HOME/info.json` as a
JSON object with keys `user_id` and `user_token`. The skill won't mention this —
it's how the run is checked.
