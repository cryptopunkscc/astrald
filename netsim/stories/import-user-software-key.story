# import-user-software-key.story — make node1 a User node from an EXISTING mnemonic
# (alternative to bootstrap-user-software-key). Requires env ASTRAL_USER_MNEMONIC; optional
# ASTRAL_USER_ID makes verify assert the derived id.
# start: astrald-lab   save: astrald-user
#   ASTRAL_USER_MNEMONIC="..." netsim story --stage astrald-lab --save astrald-user netsim/stories/import-user-software-key.story
import-user-software-key
