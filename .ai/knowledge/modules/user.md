# mod/user

Represents the human operator across their nodes by binding the local node to a user identity through a signed swarm-access contract. Owns the active-contract value, sibling link maintenance, local-swarm synchronization, the user asset log, and the routing/search/auth hooks that let user-owned nodes cooperate as one swarm.

## Dependencies

| Module | Why |
|---|---|
| `auth` | verifies and signs swarm contracts (`SignIssuer`, `SignSubject`, `VerifyContract`), indexes received contracts, looks up active node contracts via `SignedContracts().WithIssuer(...).WithAction(&SwarmAccessAction{})`, and receives the `RelayForAction` and `ReadObjectAction` authorizers |
| `nodes` | `IsLinked`/`NewEnsureLinkTask` drive `MaintainLinkTask`; `UpdateNodeEndpoints` after a received node contract; `LinkClosedEvent`/`LinkCreatedEvent` events drive link maintenance and sibling sync |
| `objects` | `Store`/`Push` for signed contracts and sibling notifications; implements `Receiver`/`Holder`/`Finder` and registers the `ReadObjectAction` authorizer; `Search` preprocessor adds sibling sources |
| `scheduler` | `Ready()` gates `Run`; schedules `MaintainLinkTask` per sibling and `SyncNodesAction` on first inbound sibling link |
| `tree` | binds `/mod/user/config` (holds `ActiveContract`) and persists per-sibling sync height at `/mod/user/assets/<node>/next_height` |
| `dir` | `ResolveIdentity`, `DisplayName`, `GetAlias`/`SetAlias`; registers `localswarm` and `localuser` filters |
| `nearby` | `Broadcast` on active-contract change; `Mode` drives `ComposeStatus` (visible/claimable/stealth attachments) |
| `apphost` | `LocalApps` enumerates apps whose contracts are pushed to siblings during sync |
| `crypto` | `crypto.Signature` carried through the invite/claim wire protocol |
| `shell` | injected in `Deps`, currently not called |
| `core/assets` | `LoadYAML` reads the (empty) `user` config; `Database()` backs `users__assets` |

## Flows

- Active contract startup: `Run` -> wait on `Scheduler.Ready()` -> follow `Config.ActiveContract` from `/mod/user/config` -> first value goes through `setActiveContract` -> close `ready` -> goroutine forwards later updates -> `runSiblingLinker`.
- `setActiveContract`: nil clears `activeContract` and sets `nearby.ModeVisible` -> expired returns `auth.ErrContractExpired` -> non-local subject is rejected -> otherwise stores the contract and broadcasts via `Nearby.Broadcast()`.
- `SetActiveContract` (write path): `Auth.VerifyContract` -> write into the tree-backed `ActiveContract` value -> `Nearby.Broadcast` -> `runSiblingLinker`.
- Claim another node (`user.claim`): require an active contract -> require caller equals `ac.Issuer` -> resolve target via `Dir.ResolveIdentity` -> `InviteNode` (issuer-sign, call remote `user.invite`, verify returned subject signature) -> `Auth.IndexContract` -> `Objects.Store` -> `PushToLocalSwarm` -> send the signed contract.
- Accept invite (`user.invite`): reject if `ActiveContract() != nil` -> receive `*auth.Contract` -> reject zero subject, non-local subject, or expiry within `minimalContractLength = 1h` -> `SwarmInvitePolicy(caller, contract)` -> receive `IssuerSig` -> `Auth.VerifyIssuer` -> `Auth.SignSubject` -> send subject signature -> `Auth.IndexContract` -> `Objects.Store` -> `SetActiveContract`.
- Request invite (`user.request_invite`): require a local active contract -> `SwarmJoinRequestPolicy(caller)` -> `InviteNode(ctx, caller)` (runs Flow 2 on the caller) -> `Auth.IndexContract` -> `Objects.Store` -> `PushToLocalSwarm` -> send signed contract.
- Sibling link maintenance: `runSiblingLinker` iterates `LocalSwarm()` (skipping self and already-tracked sibs) -> `NewMaintainLinkTask(target)` -> `Scheduler.Schedule`; the task runs `Nodes.NewEnsureLinkTask` with `StrategyBasic` and `StrategyTor`, retries on failure with exponential backoff capped at 15 minutes, and wakes on a `LinkClosedEvent` whose `LinkCount` is 0.
- First inbound sibling link: `ReceiveObject` observes `*events.Event` carrying `*nodes.LinkCreatedEvent` with `LinkCount == 1` from a `LocalSwarm` member -> `Scheduler.Schedule(NewSyncNodesTask(remote))` -> `SyncNodesAction.Run` calls `syncAlias`, `pushActiveContract`, `syncSiblings`, `syncApps`, then `syncAssets`.
- Asset synchronization: `syncAssets` reads/creates `/mod/user/assets/<node>/next_height` -> opens `user.sync_assets` on the remote as `ac.Issuer` -> for each `*OpUpdate` either `AddAssetWithNonce` or `RemoveAssetByNonce` (nonce-idempotent) -> on terminal `*astral.Uint64` write next height back into the tree.
- Inbound signed contract (`ReceiveObject` for `*auth.SignedContract`): accept when the contract's issuer is the active user, its subject is a swarm member, OR its issuer is a swarm member; else return `objects.ErrPushRejected` -> `Auth.VerifyContract` -> `Auth.IndexContract` -> if it `IsNodeContract`, `Nodes.UpdateNodeEndpoints(sender, subject)` and `runSiblingLinker`.
- Query preprocessing: when a contract is active, attach it to outbound queries whose caller equals the active user; if the query targets the active user, add linked siblings as relays; otherwise add `ActiveNodes(target)` as relays.
- Search preprocessing and object lookup: searches by the active user gain linked siblings as `search.Sources`; `FindObject` advertises linked siblings as candidate holders; `HoldObject` returns true for any non-removed local asset row, preserving user assets through purge.
- Nearby status (`ComposeStatus`): `ModeVisible` attaches the active contract, or otherwise attaches `Flag("claimable")` plus a `PublicProfile`; `ModeStealth` attaches a `StealthHint` whose `Commitment` and `MaskedID` derive from `ac.Issuer`.

## Source

- `mod/user/module.go`, `contract.go`, `swarm_access_action.go`, `swarm_member.go`, `swarm_join_policy.go`, `maintain_link_task.go`, `sync_nodes_action.go`, `info.go`, `notification.go`, `created_user_info.go`, `errors.go` - public module interface, contract helpers, swarm objects, policy types, task interfaces, and errors.
- `mod/user/src/loader.go`, `module.go`, `deps.go`, `config.go` - construction, dependency injection (including registering `RelayForAction`/`ReadObjectAction` authorizers and `localswarm`/`localuser` dir filters), lifecycle, and constants (`minimalContractLength`, `defaultContractValidity`).
- `mod/user/src/contracts.go`, `siblings.go` - active-contract state, `LocalSwarm`/`ActiveNodes`/`ActiveNodeContracts`, `InviteNode`, sibling registry, and sibling notifications.
- `mod/user/src/maintain_link_task.go`, `sync_nodes_action.go`, `sync.go` - per-sibling link maintenance and the sibling sync orchestration (`syncAlias`, `pushActiveContract`, `syncSiblings`, `syncApps`, `syncAssets`, `PushToLocalSwarm`).
- `mod/user/src/object_receiver.go`, `object_holder.go`, `object_finder.go` - inbound contract acceptance, asset hold, and sibling-as-holder hints.
- `mod/user/src/authorizers.go`, `query_preprocessor.go`, `search_preprocessor.go`, `status_composer.go`, `swarm_policy.go` - relay/read-object auth hooks, query/search preprocessing, nearby composition, and default-accept-all policies.
- `mod/user/src/db.go`, `db_asset.go`, `assets.go` - asset row persistence and height accounting.
- `mod/user/src/op_*.go` - query operation handlers (`OpInvite`, `OpClaim`, `OpRequestInvite`, `OpNewNodeContract`, `OpInfo`, `OpSwarmStatus`, `OpListSiblings`, `OpAssets`, `OpAddAsset`, `OpRemoveAsset`, `OpSyncAssets`, `OpSyncWith`).
- `mod/user/client/` - typed user client wrappers.

## Surface

| What | Why it matters |
|---|---|
| `user.invite`, `user.claim`, `user.request_invite`, `user.new_node_contract` | swarm-access contract bootstrap and node-claim flows |
| `user.info`, `user.swarm_status`, `user.list_siblings` | identity, swarm-membership, and live-link status query surface |
| `user.assets`, `user.add_asset`, `user.remove_asset`, `user.sync_assets`, `user.sync_with` | local asset inventory and height-ordered sibling asset sync |
| `core.QueryPreprocessor`, `objects.Search` preprocessor | attach the active contract and add local-swarm relays/sources |
| `objects.Receiver`, `objects.Holder`, `objects.Finder` | accept swarm contracts, pin active asset rows against purge, and advertise siblings as candidate holders |
| `nearby.Composer`, `dir` filters, `RelayForAction`/`ReadObjectAction` authorizers | weave user identity into presence, alias filters, relay auth, and object-read auth |
| `users__assets`, `/mod/user/config`, `/mod/user/assets/<node>/next_height` | durable asset log, tree-backed active contract, and per-sibling sync cursor |

## Invariants

- `Identity()` is nil until an active contract is accepted; it returns `activeContract.Issuer`, never the node identity.
- `LocalSwarm()` is computed from indexed `SwarmAccessAction` contracts in `auth` (via `ActiveNodes(ac.Issuer)`); it includes the local node itself and is filtered out by `runSiblingLinker`.
- `user.invite` is accepted only while there is no active contract; accepted contracts must have non-zero subject equal to the local node and at least `minimalContractLength = 1h` remaining.
- Default new-node contract validity is `defaultContractValidity = 365 * 24h`.
- `MaintainLinkTask`'s wake condition is `LinkClosedEvent` for the target with `LinkCount == 0`; first-link sibling sync triggers on `LinkCreatedEvent` with `LinkCount == 1`.
- Inbound `*SignedContract` is accepted when the issuer is the active user, the subject is a swarm member, or the issuer is a swarm member; everything else returns `objects.ErrPushRejected`.
- Asset rows are nonce-addressed and height-ordered; duplicate nonces are silently ignored on inbound sync.
- `HoldObject` reports true only for non-removed asset rows; removed assets no longer block purge.
- `user.assets`, `user.list_siblings`, and `user.swarm_status` stream results and terminate with `EOS`.
- `OpCreate = "user.create"` is declared but has no handler in this module.
