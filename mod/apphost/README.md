# apphost

apphost provides APIs for apps to connect to and interact with the node.

## Configuration

The config file for the module is `mod_apphost.yaml`.

### Listen

You can specify the endpoints on which apphost should listen for incoming app
connections:

```yaml
listen:
  - "unix:~/.config/astrald/apphost.sock"
  - "tcp:127.0.0.1:8625
```

For now `tcp` and `unix` sockets are supported.

### Default identity

Default identity is the identity that will be assumed by anonymous connections
to the apphost API. WARNING: this will allow any app to use this identity
without authentication. Use with caution.

#### Example

```yaml
default_identity: "0320b165fc799d3d3bb5bbdbe64590fdcabb52a81155f78a2216d6d6ca0894ccd9"
```

### Access tokens

You can define static access tokens for identities:

```yaml
tokens:
  mysecrettoken: demo
```

An app can use access tokens to authenticate to the apphost module. Default
apps (such as anc) will use the token from ASTRALD_APPHOST_TOKEN env variable:

```shell
$ export ASTRALD_APPHOST_TOKEN="mysecrettoken"
$ anc r test # will register test service as 'demo' identity
```

## Protocol

No documentation yet as the protocol is still unstable. All messages and
errors can be found it the `proto` package.