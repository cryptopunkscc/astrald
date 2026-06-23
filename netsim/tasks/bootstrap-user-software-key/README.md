# bootstrap-user-software-key

Mints a software User and installs an active contract on the operator node, turning it into a User-controlled node. verify proves it: acting as the persisted User, `apphost.whoami` reports the User id and `user.info` succeeds (it rejects without an active contract). Produces stage `one-node` in `astrald-lab`.
