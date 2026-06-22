# link-swarm

A netsim task that adopts the second node into the User's swarm, driven by the
Qwen Code agent on node1. It is the second half of the swarm phase: node1 is
already a User node (from [`bootstrap-user`](../bootstrap-user/README.md)); this
task brings node2 under the same User so the two share one swarm.

```
link-swarm [--vm <host>]      # default: node1 (the VM carrying Qwen)
```

It produces a new stage on top of the bootstrapped lab — its default starting
point is `astrald-user` (the `astrald-lab` stage after `bootstrap-user`):

```sh
netsim task --stage astrald-user --save astrald-swarm link-swarm
```

(Like `bootstrap-user`, it is **not** part of `lab.story`; each swarm step is an
incremental stage layered on the reusable base.)

## Execution model

Identical mechanic to `bootstrap-user`: **tiny script, thin prompt, intelligence
in the skill.** `run.sh` base64-ships [`prompt.md`](prompt.md) to node1 in a
single `netsim ssh` call and runs `qwen -y` as `tester`. The prompt is two
sentences — the operator is told it controls a User node, that another astrald
node is on the local network, and to adopt it following its **astral-agent**
skill's `node-adoption` playbook. The adopt flow (`user.adopt -target <node2>`,
reachability via the `nearby` module on the shared LAN) lives entirely in the
skill; the prompt restates none of it.

Reachability is already de-risked (see the task doc's discovery log): `nearby`
discovery works across netsim's per-VM NAT, so the agent needs no manual
`nodes.add_endpoint`.

## What `verify.sh` checks (independent, both ends)

It pulls raw JSON from both nodes and asserts **four facts on the host** that
together prove the swarm from both ends, with a **symmetric roster**:

1. **Both nodes hold an active contract issued by the same User** — `user.info`
   on node1 *and* node2 each shows `Issuer == <bootstrap User>` with `Subject ==`
   that node. node2 independently confirming the same User is the key both-ends
   proof that the adoption took.
2. **node1, as the User, lists node2 as a `Linked` sibling** (`user.swarm_status`).
3. **node2 lists node1 as a `Linked` sibling too** (`user.swarm_status`) — the
   symmetric roster delivered by astrald **#348**. `swarm_status` derives from
   node2's own active contract (not the caller), so this needs no User token. This
   guards the membership-race regression: pre-#348, node2 held only `User→node2`
   and its roster was `{node2}`, so it never recognized node1 — exactly the gap
   that blocked storing objects on a sibling (see `share-object`).
4. **A mutual authenticated link exists** — node2's `nodes.links` shows a link
   whose `RemoteIdentity` is node1.

node1 acts as the User via its persisted token; node2 answers under its node
identity (it holds the contract after the adoption, so no token is needed there).

`astral-query … -out json` emits a JSON **stream** (one object per line + an
`{"Type":"eos"}` terminator), so output is parsed line-by-line, not as one
document. Parsing/assertions run with host `python3`.

## Why not a "routed query" proof?

The first cut planned to prove routing with `astral-query <peer>:.spec`. That is
**not valid**: node introspection ops (`.spec`, `.id`, `.ping`) are served
locally and do **not** route to a sibling addressed by node-id — they fail even
on a fully formed swarm (verified live: every `<node2>:<op>` returned a routing
failure while the swarm was demonstrably linked). The earlier discovery-log
hypothesis that "swarm membership unlocks `<peer>:.spec`" is therefore
**disproven**. The contract + link + sibling triple above is the correct,
reproducible both-ends proof.

## Validated end-to-end

Run `astrald-user → astrald-swarm` (2026-06-17): the thin prompt drove the
operator to `user.adopt` node2 into the User's swarm; both nodes ended under one
User (`02ad7ef7…`) with a mutual link, and the rewritten `verify.sh` passes.
Stage `astrald-swarm` saved.
