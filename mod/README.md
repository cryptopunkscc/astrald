# Modules

Modules are optional extensions of the core functionality of `astrald`.
They are compiled into the node and have full access to the node's internals.
They should only be used for extending low-level functionality of the node.

### Core modules

| name                             | description                                              |
|:---------------------------------|:---------------------------------------------------------|
| admin                            | the admin console                                        |
| [apphost](apphost/src/README.md) | provides an interface for apps to interact with the node |
| [fwd](fwd/src/README.md)         | cross-network forwarding                                 |
| gateway                          | adds gateway functionality to the node                   |
| policy                           | policy management                                        |
| presence                         | discover other nodes in local networks                   |
| profile                          | allows nodes to exchange their profiles                  |
| reflectlink                      | provides link information to other nodes                 |
| relay                            | lets identites relay queries for other identities        |
| discovery                        | provides discovery mechanism                             |
| speedtest                        | a tool for benchmarking link speed                       |
| objects                          | provides objects APIs                                    |
| tcp                              | TCP driver                                               |
| tor                              | Tor driver                                               |

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