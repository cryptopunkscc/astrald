# Native services

Native services are astral apps that make use of the node's internals and need to be compiled into
the binary. These have unrestricted access to the process memory, so use of native services should
be minimized for security.

## appsupport

This service exposes a Unix Socket API to let local processes access the astral network.