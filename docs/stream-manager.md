# Stream Manager: Architecture and Policy Pipeline

This document explains the Stream Manager that orchestrates post-admission stream management in `mod/nodes/src`. It covers the data model, policy pipeline, conflict resolution, execution flow, defaults, and known gaps.

- Primary entrypoint: `StreamManager.Run(candidate *Stream, allStreams []*Stream) []*Stream`
- Policies live in `mod/nodes/src/policy_*.go`
- Streams are defined in `mod/nodes/src/stream.go`
- Action execution is delegated to `StreamController` (`mod/nodes/src/stream_controller.go`)


## High-level intent
After a stream is admitted, the Stream Manager evaluates a set of policies that can propose actions such as protecting or closing streams. The proposals are merged with clear precedence, then executed by a separate controller. The goal is to:
- Keep a healthy baseline of sibling connectivity
- Prefer higher quality networks
- Protect streams with recent activity
- Enforce upper bounds on outbound connections


## Data model (inputs and state)
- `StreamState`
  - `Candidate`: the newly admitted stream (can be closed by policies)
  - `AllStreams`: snapshot of all active streams (including `Candidate`)
- `Stream` (selected fields used by policies)
  - `id` (int): unique ID
  - `createdAt` (time.Time): used for oldest-first closures
  - `lastActivity` (time.Time): used by active-guard policy
  - `outbound` (bool): used by max-outbound policy
  - `conn` implementing `astral.Conn` and possibly `exonet.Conn`
  - `RemoteIdentity() *astral.Identity`
  - `Network() string`: one of `tcp`, `utp`, `gw`, `tor`, or `unknown`
- `StreamControl` (interface)
  - `CloseStream(s *Stream) error`
  - `ProtectStream(s *Stream) error`
  - `IsProtected(s *Stream) bool`
- `StreamController` (implements `StreamControl`)
  - Created fresh for each `Run` (per-run lifetime)
  - Owns a `sig.Set[int]` of protected stream IDs and executes actions
  - Methods: `ProtectStream`, `CloseStream`, `IsProtected`, `ClearProtection`, `ClearAllProtections`
- `StreamManager`
  - Owns the policy list
  - Creates a new `StreamController` for each `Run`
  - Does NOT implement `StreamControl`


## Policy pipeline
Policies are evaluated in the following order (as registered in `NewStreamManager`):
1) `SiblingGuardPolicy`
2) `ActiveStreamGuardPolicy`
3) `NetworkPreferencePolicy`
4) `MaxOutboundStreamsPolicy`

All policies are pure from the manager’s perspective: they take a `StreamState` snapshot and return a `PolicyDecision{Actions: []}`.

Note: Policies use internal defaults only (no external configuration through constructors).


## ASCII overview

Flow of a single `Run`:

  +---------------------+
  | New stream (S*)     |
  +---------------------+
             |
             v
  +--------------------------------------+
  | Snapshot: StreamState                 |
  |  - Candidate = S*                     |
  |  - AllStreams = [..., S*]             |
  +--------------------------------------+
             |
             v
  +------------------+   +---------------------+   +-------------------------+   +--------------------------+
  | SiblingGuard     |-->| ActiveStreamGuard   |-->| NetworkPreference       |-->| MaxOutboundStreams       |
  +------------------+   +---------------------+   +-------------------------+   +--------------------------+
             \                |                        |                                 |
              \_______________|________________________|_________________________________/
                              collects PolicyDecisions (lists of actions)
             |
             v
  +-------------------------------+
  | Merge & Resolve Conflicts     |
  |  - Protection > Eviction      |
  +-------------------------------+
             |
             v
  +--------------------------------------------+
  | Execute Actions via StreamController (SC)   |
  |  - SC is new for each Run (per-run state)   |
  |  - SC.ProtectStream / SC.CloseStream        |
  |  - Honors IsProtected                       |
  +--------------------------------------------+
             |
             v
  +-------------------------------+
  | Return closed streams []S     |
  +-------------------------------+


## Policies (current behavior)

- SiblingGuardPolicy (`policy_sibling_guard.go`)
  - Intent: Ensure at least a minimum number of links (unique remote identities) are kept.
  - Behavior: Groups streams by remote identity; when the number of unique links is at or below the minimum, it protects exactly one stream per identity (choosing the most recently active stream in each group).
  - Default: 3 links (internal constant).
  - Effect: While unique link count is ≤ 3, each identity keeps one protected stream; other streams may still be closed by later policies unless protected by other rules.

- ActiveStreamGuardPolicy (`policy_active_guard.go`)
  - Protect any stream with recent activity.
  - Default window: 5 minutes (internal constant).

- NetworkPreferencePolicy (`policy_network_preference.go`)
  - Group by `RemoteIdentity()`.
  - Sort streams by network priority: `tcp`(1) > `utp`(2) > `gw`(3) > `tor`(4) > unknown(999).
  - Keep the best-network stream for each identity; close streams on inferior networks.
  - Note: Multiple streams on the same best network are kept; only worse networks are closed.

- MaxOutboundStreamsPolicy (`policy_max_outbound.go`)
  - If the number of outbound streams exceeds the limit, close the oldest outbound streams until within limit.
  - Default limit: 10 (internal constant).
  - Sorting key: `createdAt` (oldest first).


## Conflict resolution (merge phase)

- Collect all proposed actions across policies.
- Deduplicate by stream `id`.
- Aggregate and dedupe reasons across proposals per stream.
- Apply precedence: Protection overrides eviction.

ASCII set view:

  Proposals:
    P = { stream IDs to protect }
    E = { stream IDs to evict/close }

  Resolution:
    E' = E \ P
    Final actions = Protect(P) ∪ Close(E')

This ensures any protected stream is not closed by another policy in the same run. When multiple policies propose the same action for the same stream, their reasons are merged and de-duplicated.


## Execution semantics

- Each action is executed via the `StreamController` that implements `StreamControl`.
- The controller is created per-run, so protection marks are ephemeral and reset every `Run`.
- `ProtectStream` marks the stream ID in a per-run `protectedStreams` and logs (via the manager) with reasons.
- `CloseStream` checks protection; if protected, it skips closing; otherwise it closes and logs (via the manager) with reasons.
- `Run` returns the list of streams that were actually closed as a result of executing the merged actions.


## Reason tracking

- Policies attach their policy name as a reason when proposing actions.
- During the merge phase, reasons are aggregated per stream and de-duplicated while preserving order.
- Manager logs include the merged reasons for every protect/close action.
- Note: If a stream is both protected and proposed for closure by different policies, protection wins; only protection is executed and logged. The close reasons are not logged in that case.


## Logging

- Proposals per policy: `peers.log.Logv(2, "%s proposed %d actions", policy.Name(), count)`
- Protect execution (with reasons): `peers.log.Logv(2, "protecting stream %d to %v (%v) reasons=[%s]", id, remoteID, network, reasons)`
- Close execution (with reasons): `peers.log.Infov(1, "closing stream %d to %v (%v) reasons=[%s]", id, remoteID, network, reasons)`
- Summary: `peers.log.Logv(1, "closed %d streams due to policies", n)`


## Quick glossary

- Protect: Mark a stream immune to eviction for the remainder of the current `Run`.
- Evict/Close: Close a stream unless it’s protected.
- Candidate: The newly admitted stream for which the policy pipeline is being run.
- Sibling: A peer of special interest (currently treated as any stream; future code may narrow this).
