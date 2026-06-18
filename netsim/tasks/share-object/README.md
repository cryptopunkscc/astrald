# share-object

A netsim task that stores an astral **object** on node1 and proves a swarm
sibling (node2) can obtain it **across the swarm** by its Object ID. It is the
first scenario *past* swarm formation: where `bootstrap-user` + `link-swarm`
build the swarm, `share-object` makes the swarm carry data — astral's core act,
"Identities exchanging Objects."

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

The split is deliberate and is the heart of this task's design:

* **The agent does only the STORE half.** The prompt tells the operator — acting
  as its User — to store a short, distinctive text payload as an astral object
  via the objects protocol (its **astral-agent** skill navigates
  `protocols/README → objects/README → objects.store`), and to record the
  returned Object ID. The store path is reliably reachable from a thin prompt.
* **The cross-swarm FETCH lives in `verify.sh`, not the prompt.** A thin prompt
  names no target, and the skill's rule is *"default Query target is the local
  Node; set an explicit target only when the user names one"* — so an
  agent-driven fetch would be non-deterministic. Putting the fetch in `verify.sh`
  lets it address node2 deterministically and keeps the cross-swarm assertion
  independent of agent behaviour.

The agent writes its artifacts under `~tester/.netsim/`:

| File | Purpose |
|---|---|
| `object.id` | the `data1…` Object ID returned by `objects.store` (the handle `verify.sh` fetches by) |
| `object.payload` | the exact bytes the agent stored (`verify.sh` compares node2's fetched bytes against this) |
| `share-object.log` | the agent's run log |

The store itself is, per the docs:
`echo '{"Type":"string8","Object":"<payload>"}' | astral-query objects.store -in json -out json`
→ `{"Type":"object_id.sha256","Object":"data1…"}`, run under the User token
(`ASTRALD_APPHOST_TOKEN` from `~/.netsim/user.token`) so the object lands in
node1's write-default repo as the User.

## What `verify.sh` checks (independent, both ends)

It reads the id + payload the agent persisted on node1, resolves node1's node
identity host-side (the `Subject` of node1's active contract, cross-checked
against node2's `nodes.links` `RemoteIdentity`), and then tries to pull that
exact id **from node2's vantage**, asserting the bytes match. node1 acts as the
User via its token; node2 answers under its node identity (anonymous apphost
caller, no token — exactly like `link-swarm`'s node2-side checks).

**The cross-swarm hop is inferred from the docs, not demonstrated by them.** The
astral-docs describe a `network` zone and a finder/provider layer, but contain no
worked example of one swarm member reading another's object by id. So — exactly
as `link-swarm` discovered that `<peer>:.spec` does **not** route —
`verify.sh` probes a **ladder** on node2 and reports which hop routes:

1. **explicit target** `astral-query <node1-id>:objects.load -id <ID> -out json`
   — query-target routing over the swarm link. **Primary**, because it does *not*
   depend on node2's network zone: an anonymous apphost caller has `ZoneNetwork`
   **stripped**, so it can't resolve a remote provider by zone, but it *can*
   address node1 directly and let node1 serve the read.
2. **transparent** `astral-query objects.load -id <ID> -out json` — relies on the
   read context's zone defaulting to all zones (incl. the network zone). Likely
   **blocked** for the anonymous host-side caller; kept as a bonus probe.
3. **provider find** `astral-query objects.find -id <ID> -out json` — returns
   provider **identities**, not bytes. If *only* this works, discovery crosses
   the swarm but the byte read does not — a partial finding, not a pass.

Before the read it runs a locality pre-check (`objects.contains -repo local` on
node2) so a pass reflects a genuine remote pull, not a coincidental local copy
(advisory — `objects.contains` is probabilistic, so it warns, never hard-fails).
It also separates an **authorization** rejection (`mod.objects.read_object_action`
denies the read) from a **routing** failure — different findings, never
conflated.

**PASS** iff node2 obtained the exact stored bytes for the agent-reported id
across the swarm (hop 1 or 2). `astral-query … -out json` emits a JSON **stream**
(one object/line + an `{"Type":"eos"}` terminator), parsed line-by-line with host
`python3`.

## Memory repository — a separate task, by decision

The docs expose a `memory` repository group (`objects.new_mem`, the `mem0`/
`memory` repos) — an in-memory, **non-default** write target. `share-object`
deliberately does **not** use it: it must test the *default* cross-swarm path
(node1's standard write-default repo, node2 pulls), and routing an object through
a memory repo would muddy whether a *default-repo* object crosses the swarm.
Ephemeral / `objects.new_mem` behaviour deserves its own focused task layered on
`astrald-shared` later (captured in Triage).

## Skill gap this scenario exercises

The skill has playbooks only for `swarm-management` (`node-setup`,
`node-claiming`) — there is **no objects storage/transfer playbook**. The store
half is reachable from the protocol docs alone (a real test of "are the docs
sufficient without a playbook?"); the transfer half having no playbook is itself
a finding. If the thin store prompt proves shaky live, the remedy is a small
`objects` playbook in the skill, not a fatter prompt.

## Not yet run end-to-end

Syntax-clean and registered via `link.sh` (`netsim tasks` lists `share-object`).
The harness mechanics mirror the validated `bootstrap-user` / `link-swarm`
tasks, and the `objects.store` form is taken verbatim from `objects.store.md`.
Open **CONFIRM** items, all to be pinned on the first live
`astrald-swarm → astrald-shared` run (the cross-swarm read is the inferred part):

* Whether `astral-query <node1-id>:objects.load` actually **routes** the read to
  node1 across the swarm and returns the bytes — or hits the same wall as
  `<peer>:.spec`. This is *the* unknown the run resolves.
* Whether an **anonymous** host-side caller on node2 is **authorized** to read
  node1's object (`mod.objects.read_object_action` default policy in a one-User
  swarm). If reads are denied, the likely fix is to mint a node2-side User token
  (node2 holds the contract) and read as the User, or to move the fetch into the
  operator on node1 — noted as the fallback lever, not baked into v1.
* Whether the transparent (no-target) read is genuinely blocked by the stripped
  network zone, confirming the explicit-target form as the only working hop.
* The exact `objects.load` / `objects.find` / `objects.contains` stream shapes and
  any `error_message` framing on this build (the parser is defensive but
  unverified against live output).
* Whether the thin store prompt reliably drives the operator to `objects.store`
  under the User token (vs an anonymous/local context) and to write both artifact
  files.

If the read does not route but `objects.find` does, that is a real discovery to
record (provider discovery crosses the swarm; byte read does not) — `verify.sh`
already detects and reports exactly that, the same way `link-swarm` reported the
`.spec` non-routing.
