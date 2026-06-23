# import-user-software-key

Makes the target node a User node from an existing software User, deriving the key from the BIP-39 mnemonic in `prompt.md` rather than minting fresh entropy. Verify asserts `apphost.whoami` reports that User id and `user.info` finds an active contract; if `ASTRAL_USER_ID` is set, the derived id must equal it. Drop-in alternative to `bootstrap-user-software-key`; produces stage `one-node`.
