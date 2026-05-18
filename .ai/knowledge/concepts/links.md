# Links, Sessions, and Link Establishment

## Link

* A `Link` is one authenticated brontide-encrypted connection between two
  `Identity` values.
* At most one `Link` is in active use per peer pair.
* `Link` implements `Router` and exposes `LocalIdentity()`,
  `RemoteIdentity()`, `Close()`, and `Done()`.
* A `Link` runs a continuous ping loop, tracks cumulative throughput, and may
  hold a `LinkPressureDetector`.
* `LinkCreatedEvent` and `LinkClosedEvent` report link lifecycle changes.

## Session

* A `Session` is a logical pipe over a `Link` for one routed `Query`.
* Each `Session` is tracked by a 64-bit `Nonce`.
* Read and write buffers are independent and have independent flow control.
* The default read buffer is 4 MB.
* The maximum `Data` payload is 8 KB.
* `Write` blocks when `wsize == 0`.
* `Read` frames restore credit as the receiver drains its buffer.
* A `Session` can migrate to another `Link` while open.
* During migration, `Write` blocks in `stateMigrating` until
  `CompleteMigration` switches the carrier.

## Link Lookup

* `LinkPool` returns an existing `Link` or registers a `linkWatcher`.
* `notifyLinkWatchers` delivers the new link to all pending waiters
  simultaneously.
* `RetrieveLink` accepts `WithStrategies(names...)` to limit which strategies
  are tried.
* `RetrieveLink` accepts `WithForceNew()` to bypass the existing-link check.

## Strategy Activation

* `NodeLinker` holds strategy instances for one target.
* `Activate` signals each named strategy in parallel.
* `Activate` returns a channel that closes when all strategies finish.
* `LinkStrategy` drives one link-establishment attempt.
* `LinkStrategy` exposes `Name()`, `Signal(ctx)`, and `Done()`.
* Built-in strategies:
  * `BasicLinkStrategy`: parallel direct dial across known endpoints.
  * `NATLinkStrategy`: UDP hole-punch to KCP.
  * `TorLinkStrategy`.
* `StrategyFactory` constructs a `LinkStrategy` for a target.
* `RegisterLinkStrategy` registers strategy factories per network.

## Link Pressure

* Link pressure starts when a `Link` score reaches a transport-specific
  threshold.
* The score is weighted token-bucket level plus RTT EMA.
* `LinkPressureDetector` tracks pressure state.
* The Enter threshold is higher than the Exit threshold.
* Pressure stays active until the score falls below the Exit level.
* Built-in configs:
  * `DefaultLinkPressureConfig` for TCP/KCP.
  * `TorLinkPressureConfig`.
* `LinkPressureEvent` reports pressure state changes.

## Connectivity Upgrade

On `LinkPressureEvent`:

* Prefer an existing lower-pressure link to the same peer.
* If none exists, call `RetrieveLink` with `WithForceNew()` and `StrategyNAT`.
* Call `migrateSessions` for eligible sessions.
* A session is eligible at age >= 30 s or bytes >= 1 MB.
* `sig.Switch` allows one upgrade at a time per peer.
* Each attempt enforces a 5-minute cooldown.
