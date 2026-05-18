# mod/user

Represents the human operator across their nodes by binding the local node to a user identity with a signed swarm-access contract. Owns active-contract state, sibling link maintenance, local-swarm synchronization, user asset rows, and routing/search/auth hooks that let user-owned nodes cooperate.

## Dependencies

| Module | Why |
|---|---|
| `auth` | verifies and signs swarm contracts, indexes accepted contracts, searches active node contracts, and registers relay/read authorizers |
| `nodes` | checks link state, schedules ensure-link tasks, updates node endpoints after received contracts, and handles link events |
| `objects` | stores and pushes signed contracts, sends sibling notifications, registers receiver/holder/finder hooks, and exposes read actions for authorization |
| `scheduler` | gates `Run` on `Ready`, schedules `MaintainLinkTask`, and schedules `SyncNodesAction` on first inbound sibling links |
| `tree` | binds `/mod/user/config` and persists per-sibling asset sync heights under `/mod/user/assets/<node>/next_height` |
| `dir` | resolves target aliases, reads and writes user and node aliases, and registers `localswarm` and `localuser` filters |
| `nearby` | broadcasts active-contract changes, controls claimable visibility, and receives composed visible or stealth status attachments |
| `apphost` | `syncApps` enumerates local apps and pushes their signed contracts to siblings |
| `crypto` | invite flow exchanges `crypto.Signature` objects |
| `shell` | injected into `Deps`; currently retained without direct calls |
| `core/assets` | `LoadYAML` reads config and `Database()` backs the `users__assets` table |

## Flows

- Active contract startup: `Run` waits for scheduler readiness -> follows `Config.ActiveContract` from `/mod/user/config` -> applies the first value with `setActiveContract` -> closes `ready` -> watches later updates -> runs sibling linker.
- Active contract acceptance: nil clears user identity and sets nearby visible mode -> expired contracts fail -> subject must equal the local node identity -> active contract updates `mod.activeContract` and broadcasts nearby status.
- Claim another node: `user.claim` requires an active contract and active-user caller -> resolves target -> `InviteNode` builds and issuer-signs a node contract -> remote `user.invite` returns subject signature -> verify subject -> index and store contract -> push to local swarm -> return the signed contract.
- Accept invite: `user.invite` rejects when an active contract already exists -> receives contract -> requires local node as subject and minimum validity -> applies `SwarmInvitePolicy` -> verifies issuer signature -> signs as subject -> indexes and stores -> sets active contract.
- Request invite: `user.request_invite` requires local active contract -> applies `SwarmJoinRequestPolicy` to caller -> invites caller as a new node -> indexes and stores the contract -> pushes it to local swarm.
- Sibling link maintenance: `runSiblingLinker` schedules one `MaintainLinkTask` for each non-self local swarm node not already tracked -> task runs `Nodes.NewEnsureLinkTask` with basic and Tor strategies -> retries with backoff -> wakes again on `LinkClosedEvent` with zero remaining links.
- First inbound sibling link: `ReceiveObject` observes a `LinkCreatedEvent` with `LinkCount == 1` from a local-swarm member -> schedules `SyncNodesAction` -> sync action pulls alias info, pushes active contract, pushes sibling contracts, pushes app contracts, and synchronizes assets.
- Asset synchronization: `syncAssets` reads next height from tree -> queries remote `user.sync_assets` as the active user -> applies `OpUpdate` rows with nonce idempotency -> writes returned next height back to tree.
- Inbound signed contract: `ReceiveObject` accepts only contracts issued by the active user or involving local-swarm members -> verifies and indexes the contract -> if it is a node contract, updates endpoints and reruns sibling linker.
- Query preprocessing: active-user caller gets the active contract attached -> queries targeting the active user get linked siblings as relays -> other targets get active nodes for that target as relays.
- Search preprocessing and object lookup: active-user object searches add linked siblings as sources; object finder returns linked siblings as candidate holders; holder reports true for non-removed asset rows.
- Nearby status: visible mode attaches the active contract, or claimable flag and public profile when unclaimed; stealth mode attaches a commitment and masked node identity derived from the active user.

## Source

- `mod/user/module.go`, `contract.go`, `info.go`, `notification.go`, `swarm_member.go`, `swarm_access_action.go`, `swarm_join_policy.go`, `maintain_link_task.go`, `sync_nodes_action.go`, `errors.go` - public module interface, contract helpers, query objects, policies, task interfaces, and errors.
- `mod/user/src/loader.go`, `module.go`, `deps.go`, `config.go` - construction, active-contract config binding, dependency registration, lifecycle, and constants.
- `mod/user/src/contracts.go`, `siblings.go` - active-contract state, local-swarm queries, node invitation, sibling registry, and sibling notifications.
- `mod/user/src/maintain_link_task.go`, `sync_nodes_action.go`, `sync.go` - sibling link maintenance and sync orchestration.
- `mod/user/src/object_receiver.go`, `object_holder.go`, `object_finder.go` - objects-module receiver, holder, and finder hooks.
- `mod/user/src/authorizers.go`, `query_preprocessor.go`, `search_preprocessor.go`, `status_composer.go`, `swarm_policy.go` - auth, routing, search, nearby, and policy hooks.
- `mod/user/src/db.go`, `db_asset.go`, `assets.go` - asset row persistence, height accounting, and asset notifications.
- `mod/user/src/op_*.go` - query operation handlers.
- `mod/user/client/` - typed user client wrappers.

## Surface

| What | Why it matters |
|---|---|
| `user.invite`, `user.claim`, `user.request_invite`, `user.new_node_contract` | node-claim and swarm-access contract flows |
| `user.info`, `user.swarm_status`, `user.list_siblings` | identity and sibling status query surface |
| `user.assets`, `user.add_asset`, `user.remove_asset`, `user.sync_assets`, `user.sync_with` | local asset inventory and sibling asset synchronization |
| `core.QueryPreprocessor`, `objects.Search` preprocessor | attaches user contracts and adds local-swarm relay/search sources |
| `objects.Receiver`, `objects.Holder`, `objects.Finder` | accepts swarm contracts and advertises local or sibling object holdings |
| `nearby.Composer`, `dir` filters, auth authorizers | integrates user identity with presence, alias filters, relay auth, and object-read auth |
| `users__assets`, `/mod/user/config` | durable asset log and tree-backed active contract |

## Invariants

- `Identity()` is nil until an active contract is accepted; it returns the contract issuer, not the node identity.
- `LocalSwarm()` is derived from indexed `SwarmAccessAction` contracts in `auth`; it is not the live link set.
- `user.invite` accepts only while there is no active contract.
- Accepted active contracts must be unexpired and have the local node identity as subject.
- `minimalContractLength` is 1 hour for invite acceptance; default new-node contract validity is 365 days.
- One `MaintainLinkTask` is tracked per non-self sibling in `sibs`.
- Asset rows are nonce-addressed and height-ordered; duplicate nonces from sync are ignored.
- `user.assets`, `user.list_siblings`, and `user.swarm_status` stream results and terminate with `EOS`.
