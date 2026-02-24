## Testing streams over NAT traversal (KCP)

This document describes how to establish a direct stream between two Astral nodes using a **NAT-traversed UDP path** and the **KCP transport**.

The process consists of:

* Discovering a viable NAT-traversed port pair
* Handing that path over to the transport layer
* Creating a logical Astral stream over KCP

At present, **only KCP is supported** as a transport for NAT-traversed streams.

---

## Procedure

### 1. Create a NAT traversal

Initiates NAT probing and keep-alive traffic toward the target node.

```shell
astral-query nat.new_traversal -target {identity}
```

---

### 2. List discovered NAT port pairs

Displays all currently discovered traversed port pairs.
The pair identifier (`pairId`) is used in subsequent steps.

```shell
astral-query nat.list_pairs -out json | jq
```

---

### 3. Register local ports (both nodes)

Both nodes must register the local UDP port that should be reused by the transport.
This ensures that transport traffic continues to use the same NAT mapping discovered during traversal.

```shell
astral-query kcp.set_endpoint_local_port \
  -endpoint {ip:port} \
  -local_port {localPort}
```

This step **must be performed on both nodes** before the NAT pair is taken.

---

### 4. Take the NAT pair

Stops NAT traversal keep-alives and reserves the selected port pair for transport use.

```shell
astral-query nat.pair_take -pair {pairId} -initiate true
```

The initiating node uses `-initiate true`.

---

### 5. Create an ephemeral listener (receiving node)

The receiving node creates a temporary KCP listener on the already-punched UDP port.

```shell
astral-query kcp.new_ephemeral_listener -port {port}
```

This listener must exist before the stream is created by the initiator.

---

### 6. Create a stream using an explicit transport endpoint

The initiating node creates a logical Astral stream, explicitly specifying the transport and endpoint.

```shell
astral-query nodes.new_stream \
  -target {identity} \
  -endpoint kcp:{ip:port}
```

---

## Notes

* NAT traversal is used only to **discover and preserve a viable UDP path**.
* Once a pair is taken, responsibility shifts to the transport layer.
* Local port registration is required to prevent NAT rebinding.
* Timing between steps is important, as NAT state may expire after traversal keep-alives stop.
* 
---

## Expected result

If successful:

* UDP traffic continues on the same external port pair
* Traversal keep-alives are replaced by KCP traffic
* A direct Astral stream is established between the two nodes
