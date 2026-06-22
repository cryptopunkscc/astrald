# share-object

A netsim task that has node1 (acting as its User) **store an astral object ON its
swarm sibling** (node2) and read it back, then independently proves node2
physically holds it. It is the first scenario *past* swarm formation: where
`bootstrap-user` + `link-swarm` build the swarm, `share-object` makes the swarm
carry data the hard way — one member **writing** an object onto another — astral's
core act, "Identities exchanging Objects."

```
share-object [--vm <host>]      # default: node1 (the VM carrying Qwen)
```

It produces a new stage on top of the formed swarm — its default starting point
is `astrald-swarm` (two nodes in one User Swarm):

```sh
netsim task --stage astrald-swarm --save astrald-shared share-object
```

(Like the other swarm tasks, it is **not** part of `lab.story`; each step is an
incremental stage layered on the reusable base.)

## Execution model

Same mechanic as `bootstrap-user` / `link-swarm`: **tiny script, thin prompt,
intelligence in the skill.** `run.sh` base64-ships [`prompt.md`](prompt.md) to
node1 in a single `netsim ssh` call and runs `qwen -y` as `tester`.

* **The agent stores ON the other node.** The prompt tells the operator — acting
  as its User — to store a short, distinctive text payload as an astral object
  **on the sibling**, addressing it explicitly as the query target
  (`<node2>:objects.store`), then to load it back **from that node** and confirm
  the bytes round-trip. This exercises the comprehension axis (find the sibling,
  set an explicit query target, store + load) and the newly-unblocked write
  capability in one go. It records the id, the stored payload, the read-back, and
  the node id it targeted.
* **The independent proof lives in `verify.py`.** A host-side check confirms node2
  physically holds the object, deterministically, without trusting the agent.

The agent writes its artifacts under `~tester/.netsim/`:

| File | Purpose |
|---|---|
| `object.id` | the `data1…` Object ID returned by `objects.store` |
| `object.payload` | the exact bytes the agent stored (the node2-local read is compared against this) |
| `object.readback` | the bytes the agent read back from node2 (advisory cross-check) |
| `object.target` | the node id the agent stored on (cross-checked against node2's real identity) |
| `share-object.log` | the agent's run log |

The store is, per the docs, the same `string8` form as a local store but with an
explicit target:
`echo '{"Type":"string8","Object":"<payload>"}' | astral-query <node2-id>:objects.store -in json -out json`
run under the User token (`ASTRALD_APPHOST_TOKEN` from `~/.netsim/user.token`); the
read-back is `astral-query <node2-id>:objects.load -id <id>`.

## What `verify.py` checks (independent, decisive)

`verify.py` does not trust `run.sh` or the agent's read-back. It reads the id +
payload the agent persisted on node1 and proves **node2 physically holds the
object in its local repo**:

* `objects.store` writes to `WriteDefault()` — the **`local`** repo — so the object
  lands in node2's `local` repo. `verify.py` reads it straight back from there, on
  node2:
  * `astral-query objects.load -id <ID> -repo local` → bytes must equal the stored
    payload — **the decisive check**;
  * `astral-query objects.contains -repo local -id <ID>` → corroborating bool.
* Both ops are **ungated and repo-pinned**, so a successful repo-local load on node2
  is conclusive: the bytes came from node2's own storage, not a network re-fetch
  from node1. node2 answers under its node identity (anonymous host-side caller, no
  token — repo-local load/contains need no authorization).

It also resolves node2's real identity host-side (the `Subject` of node2's active
contract, with node1's `nodes.links` `RemoteIdentity` as a fallback) to cross-check
the node the agent claims it targeted, and notes (advisory) whether node1 also
holds a copy. **PASS** iff node2's `local` repo returns the exact stored bytes.
`astral-query … -out json` emits a JSON **stream** (one object/line + an
`{"Type":"eos"}` terminator), parsed line-by-line with host `python3`.

## Why storing on a sibling works now

Earlier runs hit `query rejected (1)` on `<node2>:objects.store` — node2 refused to
relay node1's User-authenticated query because its swarm roster was **asymmetric**:
after `link-swarm`, node2 held only `User→node2` and never learned `User→node1`, so
node1 was absent from node2's `LocalSwarm()` and `AuthorizeRelayFor` denied the
relay before the query reached any objects op.

astrald **#348** ("Sync full swarm roster to a newly invited node", on `master`)
fixes this: both invite paths now schedule `SyncNodesAction` against the joined
node right after indexing, so node2 converges to the full, symmetric roster
(including `User→node1`). node2 now recognizes node1 as a sibling →
`AuthorizeRelayFor` allows the relay → the query reaches `op_store`, which has **no
auth gate**, and the write lands.

> **Caveat (recorded, not a blocker):** `objects.CreateObjectAction` is still an
> unwired stub — `op_store` performs **no** authorization. So a cross-swarm store
> works but is *unauthenticated at the op level* (any caller that can route + relay
> can write). Hardening that (wire `CreateObjectAction` + an `AuthorizeCreateObject`
> grant) is a separate, design-gated task ("Wire up the object-creation
> authorizer"). This scenario tests the *functional* write path, not the
> authorization model.
>
> **Pending the first live `astrald-swarm → astrald-shared` run:** this is verified
> at the code level on `master`; confirm end-to-end live, and confirm the thin
> prompt reliably drives the agent to resolve the sibling and set an explicit query
> target.

## Memory repository — a separate task, by decision

The docs expose a `memory` repository group (`objects.new_mem`, the `mem0`/
`memory` repos) — an in-memory, **non-default** write target. `share-object`
deliberately does **not** use it: it must test the *default* write path (the
sibling's standard write-default `local` repo). Ephemeral / `objects.new_mem`
behaviour deserves its own focused task layered on `astrald-shared` later (captured
in Triage).

## Skill gap this scenario exercises

The skill has playbooks only for `swarm-management` (`node-setup`,
`node-adoption`) — there is **no objects storage/transfer playbook**. Storing on a
sibling (resolve the sibling, set an explicit query target, `objects.store` +
`objects.load`) must be reached from the protocol docs alone — a real test of "are
the docs sufficient without a playbook?". If the thin prompt proves shaky live, the
remedy is a small `objects` playbook in the skill, not a fatter prompt.
