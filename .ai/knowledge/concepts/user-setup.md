# User Node Setup

How a node binds itself to a user identity by obtaining an active `SignedContract` that grants `SwarmAccess`, and how additional nodes join the same user's swarm.

## Scope

* A node operates under `Config.ActiveContract`, a `tree.Value[*auth.SignedContract]` bound at `/mod/user/config`.
* `Module.Identity()` returns `activeContract.Issuer` (the user) and is nil until a contract is active.
* `user.NewNodeContract(issuer, subject, duration)` builds an unsigned `*auth.Contract` whose only permit is `SwarmAccessAction.ObjectType()`.
* A contract is a "node contract" iff `IsNodeContract(c)` is true (i.e. it has at least one `SwarmAccessAction` permit).

## Key Types

| Type | Role |
|------|------|
| `auth.Contract` | Unsigned `{Issuer, Subject, Permits, ExpiresAt}` |
| `auth.SignedContract` | Contract plus `IssuerSig` and `SubjectSig` (`*crypto.Signature`) |
| `user.SwarmAccessAction` | Permit that makes a contract a node contract |
| `Config.ActiveContract` | `tree.Value[*auth.SignedContract]` persisted under `/mod/user/config` |

## Invariants

* A claimed node rejects another `user.invite` (code 2).
* A node without an active contract cannot issue or relay invites.
* Only the user identity (`ac.Issuer`) may call `user.claim` on an active node (code 3 otherwise).
* The accepted contract's `Subject` must equal the local node identity and have at least `minimalContractLength = 1h` remaining.
* The receiving node verifies the full `SignedContract` (issuer + subject signatures) before indexing and storing it.
* `SwarmInvitePolicy` and `SwarmJoinRequestPolicy` decide acceptance; both defaults accept everything.

## Flow 1: First-Node Bootstrap

When there is no existing swarm yet, one node holds both the user key and the node key locally.

```text
1. user.new_node_contract  ->  *auth.Contract{Issuer=user, Subject=node, SwarmAccess permit}
2. auth.sign_contract      ->  *auth.SignedContract signed as issuer and subject
3. write the SignedContract into Config.ActiveContract (tree.Value)
```

Step 3 may run internally (writing directly into the `tree.Value`) or via `Module.SetActiveContract`. Either way the follower in `Run` calls `setActiveContract`, which becomes the new `Module.Identity()` and triggers `Nearby.Broadcast` and `runSiblingLinker`.

## Flow 2: Invite From An Active Node

When the user already controls one node and wants to add an unclaimed one. The initiator already has an active contract and calls `user.claim` with `Target=<nodeAlias>`.

### `OpClaim` on the initiator (active node)

1. `ActiveContract() == nil` -> `RejectWithCode(2)`.
2. `q.Caller() != ac.Issuer` -> `RejectWithCode(3)`.
3. `Dir.ResolveIdentity(target)` -> node identity.
4. `InviteNode(ctx, nodeID)` (see below).
5. `Auth.IndexContract(signed)`.
6. `Objects.Store(WriteDefault(), signed)`.
7. `go PushToLocalSwarm(signed)`.
8. Send the `SignedContract` back to the caller.

### `InviteNode` (internal helper on the active node)

1. `user.NewNodeContract(ac.Issuer, nodeID, defaultContractValidity = 365*24h)`.
2. `Auth.SignIssuer(ctx, signed)` -> `issuerSig`.
3. `userClient.New(nodeID).Invite(ctx, contract, issuerSig)` opens `user.invite` on the target.
4. Receive `subjectSig` over the channel.
5. `Auth.VerifySubject(signed)`.
6. Return the assembled `SignedContract`.

### `OpInvite` on the target (unclaimed node)

Called by the initiator. Wire protocol:

```text
<- *auth.Contract            (from inviter)
<- *crypto.Signature         (IssuerSig)
-> *crypto.Signature         (SubjectSig: node signs the contract)
```

1. `ActiveContract() != nil` -> `RejectWithCode(2)`.
2. Receive `*auth.Contract`. Reject when `Subject.IsZero()`, when the local node is not the subject, or when expiry minus now is less than `minimalContractLength`.
3. `SwarmInvitePolicy(caller, contract)` -> on rejection send `ErrInvitationDeclined`.
4. Receive `IssuerSig`. `Auth.VerifyIssuer(signed)`.
5. `Auth.SignSubject(ctx, signed)` -> send `subjectSig`.
6. `Auth.IndexContract(ctx, signed)`.
7. `Objects.Store(WriteDefault(), signed)`.
8. `SetActiveContract(signed)`, which verifies the contract, writes it into the tree, calls `Nearby.Broadcast`, and runs `runSiblingLinker`.

## Flow 3: Request Invite

When an unclaimed node asks to join a swarm. The requester calls `user.request_invite` on any node that already has an active contract.

### `OpRequestInvite` on the swarm node

1. `ActiveContract() == nil` -> `RejectWithCode(2)`.
2. `SwarmJoinRequestPolicy(q.Caller())` -> on rejection send `ErrRequestDeclined`.
3. `InviteNode(ctx, q.Caller())`; the rest of Flow 2 runs against the requester's `OpInvite`.
4. `Auth.IndexContract(signed)`.
5. `Objects.Store(WriteDefault(), signed)`.
6. `go PushToLocalSwarm(signed)`.
7. Send the `SignedContract` back to the requester.

The requester's own `OpInvite` is the one that finishes signing and activates the contract on its side.

## After Setup

* Target node: `activeContract` is set, `Identity()` returns the user identity, `Nearby.Broadcast` fires, and `runSiblingLinker` schedules `MaintainLinkTask` for every other node in `LocalSwarm()`.
* Swarm visibility: `auth` indexes the `SignedContract` and `LocalSwarm()` now includes the new node (computed from `Auth.SignedContracts().WithIssuer(user).WithAction(&SwarmAccessAction{}).Find`).
* Object propagation: `Objects.Push` distributes the contract to siblings; on the receiving side `ReceiveObject` accepts contracts whose issuer is the active user, whose subject is a swarm member, or whose issuer is a swarm member, then `Auth.IndexContract`s them and reruns its own `runSiblingLinker`.

## Signing Split

| Who signs | Field | Method |
|-----------|-------|--------|
| User (issuer) | `IssuerSig` | `Auth.SignIssuer` |
| Node (subject) | `SubjectSig` | `Auth.SignSubject` |

`signAs` tries `Crypto.ObjectSigner` (ASN1) first and falls back to `Crypto.TextObjectSigner` (BIP137); `VerifyContract` dispatches per `Signature.Scheme`. `auth.sign_contract` (`OpSignContract`) runs both sides in one call and is what Flow 1 uses when both keys are local.

## Policy Hooks

| Policy | Default | Applied in |
|--------|---------|------------|
| `SwarmInvitePolicy(invitee, contract) bool` | accept all | `OpInvite`: target decides whether to accept an invitation |
| `SwarmJoinRequestPolicy(requester) bool` | accept all | `OpRequestInvite`: swarm node decides whether to honor a join request |

Both defaults live in `swarm_policy.go` (`SwarmInviteAcceptAll`, `SwarmJoinRequestAcceptAll`) and are returned by `GetSwarmInvitePolicy` / `GetSwarmJoinRequestPolicy`.
