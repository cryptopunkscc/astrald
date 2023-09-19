# TCP tunnels guide

## Overview

TCP tunnels are quite easy to set up using the built-in `fwd` module.
The idea is simple: one node exports locally accessible TCP addresses as
a service, and other nodes import these services by binding them to their
local TCP ports.

## Example

In this example we will export SSH and Bitcoin services from node `demo` and
access it from a node `tester`.

First, on node `demo`, create a `mod_fwd.yaml` file in your
[config directory](quickstart.md#config-directory) with the following content
and restart the node.

```yaml
forwards:
  "astral://ssh": "tcp://127.0.0.1:22"
  "astral://bitcoin": "tcp://127.0.0.1:8333"
  "astral://bitcoin-rpc": "tcp://127.0.0.1:8332"
```

This just tells the `fwd` module to register `ssh`, `bitcoin` and
`bitcoin-rpc` services on the node and simply forward all incoming queries
to respective TCP addresses.

You can use `anc` tool to see if the services are exported correctly by
querying the ssh service:

```shell
$ anc q ssh
connected.
SSH-2.0-OpenSSH_8.9p1 Ubuntu-3ubuntu0.4
```

Now we can move on to importing these services on the other node.
On node `tester`, create a `mod_fwd.yaml` file with the following content.
Remember to restart your node afterwards.

```yaml
forwards:
  "tcp://127.0.0.2:8022": "astral://demo:ssh"
  "tcp://127.0.0.2:8333": "astral://demo:bitcoin"
  "tcp://127.0.0.2:8332": "astral://demo:bitcoin-rpc"
```

This binds local TCP ports on a loopback IP 127.0.0.2 to the exported services
on `demo`. You should now be able to connect to the ssh service on node `demo`
through the mapped local port:

```shell
$ ssh 127.0.0.2 -p 8022
```

You can make your Bitcoin node on `tester` connect only to your node running
on `demo` by adding this line to your `bitcoin.conf` file:

```text
connect=127.0.0.2:8333
```

You can also use RPC by adding the following:

```text
rpcconnect=127.0.0.2
rpcuser=myuser
rpcpassword=mypassword
```

## Advanced config

Check [fwd's documentation](../mod/fwd/README.md) for detailed options.