# import-user-software-key

Drives the Qwen operator on node1 to make it a **User-controlled node from an
existing software User** — deriving the User key from a provided BIP-39 mnemonic
(`ASTRAL_USER_MNEMONIC`) instead of minting a fresh one — following the
astral-agent skill's node-setup playbook. `verify.sh` confirms node1 answers as
that User (and, if `ASTRAL_USER_ID` is set, that the derived id matches exactly).
A drop-in alternative to `bootstrap-user-software-key`; produces stage `astrald-user`.
