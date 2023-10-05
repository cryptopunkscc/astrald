# Modules

Modules are optional extensions of the core functionality of `astrald`.
They are compiled into the node and have full access to the node's internals.
They should only be used for extending low-level functionality of the node.

### Core modules

| name                         | description                                              |
|:-----------------------------|:---------------------------------------------------------|
| admin                        | the admin console                                        |
| [apphost](apphost/README.md) | provides an interface for apps to interact with the node |
| [fwd](fwd/README.md)         | Cross-network forwarding                                 |
| gateway                      | adds gateway functionality to the node                   |
| policy                       | policy management                                        |
| presence                     | discover other nodes in local networks                   |
| profile                      | allows nodes to exchange their profiles                  |
| reflectlink                  | provides link information to other nodes                 |
| router                       | lets identites route queries via links between nodes     |
| sdp                          | provides service discovery mechanism                     |
| speedtest                    | a tool for benchmarking link speed                       |
| storage                      | provides storage APIs                                    |

### Enabled modules

By default, all compiled modules are enabled. To manually select which modules
should be enabled, add the following to `node.yaml` in your
[config directory](../docs/quickstart.md#config-directory):

```yaml
modules:
  - admin
  - apphost
  - ...
```