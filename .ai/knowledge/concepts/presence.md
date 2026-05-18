# Presence

Presence is the periodic (~5 min) signed UDP broadcast of a node's
reachability.

## Composition

* `mod/nearby` builds a `StatusMessage` by calling each registered
  `Composer`.
* `StatusMessage` is a typed bundle of attachments.
* `mod/nearby` does not know about TCP, KCP, or Tor.
* Each transport appends its own endpoint objects.
* Adding a transport means adding a composer. Discovery stays unchanged.

## Receiver

* Incoming messages are cached by source IP.
* `ResolveStatus` identifies the sender.
* Visible mode uses a signed `NodeContract` from the bundle.
* Stealth mode uses a `StealthHint`: node ID XOR user ID plus nonce.
* Only peers that know the user identity can recover the stealth sender.
* `ResolveEndpoints` reads typed endpoint objects from the same cache.
* Presence and endpoint resolution use the same data at different call sites.

## Size Limit

* Each attachment has a 4 KB cap from the UDP datagram constraint.
* Attach one address, not a routing table.

## Stealth

* Transports attach nothing.
* Only the `StealthHint` is present.
* Peers without the user identity see no endpoints.
* Peers without the user identity cannot recover the node ID.
* No hint means no broadcast.
