# User Node Setup

How a node gets a user identity by obtaining an active `SignedNodeContract`, and how more nodes join the same user's swarm.

## Scope

* A node operates under `config.ActiveContract`.
* `activeContract` is a `SignedContract` stored as a `tree.Value`.
* `mod.Identity()` returns `activeContract.Issuer`.
* `mod.Identity()` returns `nil` when no active contract is set.
* `user.NewNodeContract(issuer, subject, duration)` creates an unsigned `auth.Contract` with one `SwarmAccessAction` permit.

## Key Types

| Type | Role |
|------|------|
| `auth.Contract` | Unsigned contract: `Issuer` (user), `Subject` (node), permits (`SwarmAccessAction`), `ExpiresAt` |
| `auth.SignedContract` | Signed contract: contract + `IssuerSig` (user key) + `SubjecSig` (node key) |
| `user.SwarmAccessAction` | Permit that makes a contract a node contract |
| `activeContract` | `SignedContract` stored in `config.ActiveContract` (`tree.Value`) |

## Invariants

* A claimed node rejects a second invite.
* A node without an active contract cannot invite another node.
* Only the user identity (`ac.Issuer`) can call `user.claim` on an active node.
* The target node must be the contract `Subject`.
* The target node must accept the contract before it becomes active.
* The target verifies the full `SignedContract` before storing and indexing it.
* Policy hooks decide whether invites and join requests are accepted.

## Flow 1: First Node Bootstrap

Use this when there is no existing swarm. One node holds both the user key and the node key locally.

```text
1. user.new_node_contract  ->  unsigned Contract{Issuer=user, Subject=node}
2. auth.sign_contract      ->  SignedContract signed as issuer and subject
3. SetActiveContract(signed)
```

Step 3 runs internally, for example by writing directly to the `tree.Value`.

After step 3, the node has an `activeContract`, and `mod.Identity()` returns the user identity.

## Flow 2: Invite From User Node

Use this when the user already has one active node and wants to add an unclaimed node. The initiator already has an active contract and acts for the user.

The initiator calls `user.claim` with `target=<nodeAlias>`.

### `OpClaim` on the initiator

Caller must be the user identity.

1. Check `ActiveContract() != nil`; reject with code 2 when no active contract exists.
2. Check `q.Caller() == ac.Issuer`; reject with code 3 when the caller is not the user.
3. Resolve the target node identity.
4. Call `InviteNode(ctx, targetID)`.
5. Index and store the returned `SignedContract`.

### `InviteNode(ctx, nodeID)`

Internal helper on the active node.

1. Create `contract = NewNodeContract(ac.Issuer, nodeID, defaultContractValidity)`.
2. Sign as issuer: `issuerSig = auth.SignIssuer(ctx, contract)`.
3. Call `userClient.New(nodeID).Invite(ctx, contract, issuerSig)` to open `user.invite` on the target.
4. Receive `subjectSig` from the target.
5. Verify `subjectSig` with the node's secp256k1 key.
6. Return `SignedContract{contract, issuerSig, subjectSig}`.

### `OpInvite` on the target

Called by the initiator. Wire protocol:

```text
<- Contract                  (from inviter)
-> crypto.Signature          (subjectSig: node signs the contract hash)
<- crypto.Signature          (issuerSig: user sig sent back)
```

1. Reject with code 2 when the node already has an active contract.
2. Validate the contract: `Subject == node.Identity()` and at least `minimalContractLength = 1h` remains.
3. Consult `SwarmInvitePolicy(caller, contract)`; default accepts all.
4. Sign as subject: `subjectSig = auth.SignSubject(ctx, contract)`.
5. Send `subjectSig`.
6. Receive `issuerSig`.
7. Assemble the full `SignedContract` and call `auth.VerifyContract`.
8. Index with `auth.IndexContract`.
9. Store with `objects.Store`.
10. Call `SetActiveContract(signed)`.

After `SetActiveContract`, the node broadcasts presence with `nearby.Broadcast()` and starts `runSiblingLinker()`.

## Flow 3: Request Invite

Use this when an unclaimed node asks to join a swarm without the user initiating the invite.

The requesting node calls `user.request_invite` on any node that already has an active contract.

### `OpRequestInvite` on the existing swarm node

Caller is the requesting node.

1. Reject with code 2 when the handler node has no active contract.
2. Consult `SwarmJoinRequestPolicy(caller)`; default accepts all.
3. Call `InviteNode(ctx, q.Caller())`; the rest follows Flow 2.
4. Index and store the signed contract.
5. Send `SignedContract` back to the requester.

The requester still accepts the invitation through its own `OpInvite` handler, which `InviteNode` on the swarm node calls.

## Successful Setup

After any flow completes:

* Target node: `activeContract` is set, `Identity()` returns the user identity, and `runSiblingLinker()` runs.
* Swarm: the `auth` module indexes the `SignedContract`; `LocalSwarm()` includes the new node.
* Presence: `nearby.Broadcast()` runs so nearby peers can discover the change.
* Next sibling link: `pushActiveContract` sends the `SignedContract`; `receiveSignedContract` on the sibling re-indexes it and calls `runSiblingLinker()`.

## Signing Split

| Who signs | Signature | Method |
|-----------|-----------|--------|
| User (issuer) | `IssuerSig` | `auth.SignIssuer` signs the contract object hash (ASN1) |
| Node (subject) | `SubjecSig` | `auth.SignSubject` signs the contract object hash (ASN1) |

`auth.sign_contract` (`auth.OpSignContract`) signs as both issuer and subject in one call. Flow 1 uses it when both keys are local.

## Policy Hooks

| Policy | Default | Applied in |
|--------|---------|------------|
| `SwarmInvitePolicy(caller, contract) bool` | accept all | `OpInvite`: node decides whether to accept an invitation |
| `SwarmJoinRequestPolicy(requester) bool` | accept all | `OpRequestInvite`: swarm node decides whether to honor a join request |
