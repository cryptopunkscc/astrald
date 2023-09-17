# TCP tunnels

## Overview

TCP tunnels are quite easy to set up using the built-in `tcpfwd` module.
The idea is simple: one node exports locally accessible TCP addresses as
a service, and other nodes import these services by binding them to their
local TCP ports.

## Example

In this example we will export SSH and Bitcoin services from node A and
access it from a node B.

First, on node A, create a `mod_tcpfwd.yaml` file in your
[config directory](quickstart.md#config-directory) with the following content
and restart the node.

```yaml
out:
  ssh: "127.0.0.1:22"
  bitcoin: "127.0.0.1:8333"
  bitcoin-rpc: "127.0.0.1:8332"
```

This just tells the `tcpfwd` module to register `ssh`, `bitcoin` and
`bitcoin-rpc` services on the node and simply forward all incoming queries
to respective TCP addresses.

You can use `anc` tool to see if the services are exported correctly by querying
the ssh service:

```shell
$ anc q ssh
connected.
SSH-2.0-OpenSSH_8.9p1 Ubuntu-3ubuntu0.4
```

Now we can move on to importing these services on the other node.
On node B, create a `mod_tcpfwd.yaml` file with the following content.
Remember to replace `A` with the actual alias or public key of your node and
restart your node afterwards.

```yaml
in:
  "127.0.0.2:8022": "A:ssh"
  "127.0.0.2:8333": "A:bitcoin"
  "127.0.0.2:8332": "A:bitcoin-rpc"
```

This binds local TCP ports on a loopback IP 127.0.0.2 to the exported service
on A. You should now be able to connect to the ssh service on node A through
the mapped local port:

```shell
$ ssh 127.0.0.2 -p 8022
```

You can make your Bitcoin node on B connect only to your node running on A
by adding this line to your `bitcoin.conf` file:

```text
connect=127.0.0.2:8333
```

You can also use RPC by adding the following:

```text
rpcconnect=127.0.0.2
rpcuser=myuser
rpcpassword=mypassword
```
