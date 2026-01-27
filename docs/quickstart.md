# Quick start

## Requirements

- [Go 1.21](https://go.dev/dl/)

## Installation

```shell
$ git clone https://github.com/cryptopunkscc/astrald
$ cd astrald
$ go install ./cmd/astrald ./cmd/anc
```

This will install two binaries:
- `astrald` - the node daemon
- `anc` - a basic tool to interact with the astral network (astral netcat)
- 
> Note: `go install` puts binaries in `$HOME/go/bin` by default. Make sure to
add this directory to your $PATH.

## Running the node

Start the node:

```shell
$ astrald
(0) 16:20:33.799 - [node] astral core demo (02204655fc5085bb3a4b53aba35c105a46b89f4d81a655ee579e1aa7fe34c0059e) starting...
...
```

`demo` is the alias of your node followed by its public key in parens. The
public  key is the canonical way to represent an identity. Aliases are
assigned  locally (sort  of like /etc/hosts file) and should only be used
for convenience.

### Config directory (Linux)

By default, `astrald` will use `$HOME/.config/astrald` directory for all
resource and  config files. You can specify a different path using `-datadir`
option.

#### Config directory (MacOS)

On Mac, `astrald` will use `~/Library/Application Support/astrald` directory for resource and config files. 

## Default identity

In order to interact with the node you need to have an identity as a user.
Since you  don't have one yet, you can use your node's identity to set
things up. Create a config file called `mod_apphost.yaml` in the config
directory with the following content:

```yaml
default_identity: demo
```

Replace `demo` with the alias of your node or its public key. You can use both
aliases  and public keys in config files, but keep in mind that if you change
aliases you also need to update your config files.

Save the file and restart `astrald`. Anonymous app connections to the node will
now have node's identity. Since this lets any app use node's identity on the
network, as soon as you set up your own identity make sure to remove the
`default_identity` from the config file.

To test if everything works as expected try using `anc`:

```shell
$ anc r test
listening on test
```

If you get an `unauthorized` error, something is wrong with the config (or you
forgot to restart the node).

## Admin console

Now that you have your identity set up you can access the admin console
using `anc`:

```shell
$ anc q admin
connected.
demo@demo> 
```

The first `demo` is the identity the user is using to interact with the admin
console, the second `demo` is the identity of the node.

If you got this far, you have a fully functional astral node up and running.

## What's next?

* [TCP tunnels](tcp-tunnels.md) guide.
