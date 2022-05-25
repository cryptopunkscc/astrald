# inet

Support for IP networks.

## Presence protocol

Presence protocol is used to announce/discover astral nodes on local network(s). It's a binary protocol and uses
UDP port 8829 for all (broadcast and unicast) communication.

To announce your presence on a network, broadcast a message at most every 5 minutes. Similarly, nodes that
have not sent any messages for more than 5 minutes should be considered gone. You can ask all nodes to message
you directly by setting the `discover` flag.

### Message structure

| len | type   | name     | description          |
|-----|--------|----------|----------------------|
| 4   | uint32 | version  | always 0x61700000    |
| 33  | bytes  | identity | node identity        |
| 2   | uint16 | port     | tcp port for linking |
| 1   | uint8  | flags    | flags (see below)    |

Flags:

| flag | name     | description                                          |
|------|----------|------------------------------------------------------|
| 0x01 | discover | set this flag to discover other nodes on the network |
| 0x02 | bye      | indicate that you're about to leave the network      |

When `discover` flag is set, other nodes on the network are asked to notify you of their presence.
If a node chooses to respond, they will send back a direct message (still on the 8829 port).

The `bye` flag can be used to inform other nodes that you're leaving the network, so that they don't have to wait
for the timeout to figure out you're gone.

