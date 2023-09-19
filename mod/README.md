# Modules

Modules are optional extensions of the core functionality of `astrald`.
They are compiled into the node and have full access to the node's internals.
They should only be used for extending low-level functionality of the node.

### Core modules

| name                         | description                                              |
|:-----------------------------|:---------------------------------------------------------|
| admin                        | the admin console                                        |
| [apphost](apphost/README.md) | provides an interface for apps to interact with the node |
| discovery                    | provides service discovery mechanism                     |
| gateway                      | adds gateway functionality to the node                   |
| presence                     | discover other nodes in local networks                   |
| profile                      | allows nodes to exchange their profiles                  |
| reflectlink                  | provides link information to other nodes                 |
| route                        | lets identites route queries via links between nodes     |
| speedtest                    | a tool for benchmarking link speed                       |
| storage                      | provides storage and sharing APIs                        |
| [fwd](fwd/README.md)         | Cross-network forwarding                                 |

### Enabled modules

By default, all compiled modules are enabled. To manually select which modules
should be enabled, add the following to `node.yaml` in your
[config directory](../docs/quickstart.md#config-directory):

```yaml
modules:
  - admin
  - apphost
  - connect
  - ...
```