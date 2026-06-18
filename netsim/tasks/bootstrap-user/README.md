# bootstrap-user

A netsim task that turns the operator node into a **User-controlled node**,
driven by the Qwen Code agent running inside the VM. It is the first half of the
swarm phase: it establishes identity only; it does not link or claim anything
(that is [`link-swarm`](../link-swarm/README.md)).

```
bootstrap-user [--vm <host>]      # default: node1 (the VM carrying Qwen)
```

After it runs, the node holds an active `mod.user.swarm_access_action` contract
(issuer = a fresh software User, subject = this node), and a User-bound apphost
token is persisted so later tasks can act as the User. It produces a new stage
on top of the lab base — run it standalone against `astrald-lab` with:

```sh
netsim task --stage astrald-lab --save astrald-user bootstrap-user
```

(`bootstrap-user` is deliberately *not* part of `lab.story`: `astrald-lab` stays
the reusable base, and each swarm step is an incremental stage layered on it.)

## Execution model

`run.sh` runs on the host (cwd = simulation root) and does almost nothing
itself: it base64-ships [`prompt.md`](prompt.md) to the guest in a single
`netsim ssh <host> -- <one-argv>` call and invokes `qwen` as user `tester`,
non-interactively (one-shot positional prompt + `-y`), against that prompt. The
astral work — minting a software User, signing and installing the node contract,
persisting a token — is carried out by the agent, not by the script.

This is the design principle, taken one step further than the other tasks:
**tiny script, thin prompt, intelligence in the skill.** The prompt does not
spell out the contract procedure — it states the situation and the goal in plain
sentences and tells the agent to follow its **astral-agent** skill, whose
`node-setup` playbook (software-User path) is exactly this flow. The prompt
carries only what the skill cannot know: the machine-specific files `verify.sh`
will look for, idempotency, and the success criterion. If the skill is present
and sufficient, that is all the agent needs; exercising that is part of the
test.

The agent writes its artifacts under `~tester/.netsim/`:

| File | Purpose |
|---|---|
| `user.id` | the User's hex public key (the User identity) |
| `user.token` | a User-bound apphost access token (also exported in `~/.bashrc`) |
| `bootstrap-user.log` | the agent's run log |

`verify.sh` is an **independent** re-check: it reads `user.id` + `user.token`,
acts as the User, and asserts `apphost.whoami` reports the User and `user.info`
returns the active contract (which the op rejects with code `2` when absent).

## The contract flow (driven by the skill, not the prompt)

For reference only — this is what the astral-agent `node-setup` playbook
(software User) walks the agent through; the prompt does **not** restate it:

1. Mint a `secp256k1` User key via the `bip137sig` ops
   (`new_entropy → mnemonic → seed → derive_key`).
2. Store the private key via `objects.store` so `crypto` indexes it as a signer.
3. Derive the User identity (`crypto.public_key`).
4. (optional) `dir.set_alias` for a readable name.
5. `user.new_node_contract -user <user-id>` (subject defaults to this node).
6. `auth.sign_contract` — co-signs as issuer + subject (both keys are local).
7. Install at tree path `/mod/user/config/active_contract` via `tree.set`.
8. `apphost.create_token` for the User; persist + export it.
9. Confirm with `apphost.whoami` + `user.info`.

## Not yet run end-to-end

The harness mechanics mirror the validated `configure-astral-agent` task, and
the `qwen -y "<prompt>"` invocation matches what was confirmed against the live
lab. Still unverified until run on a fresh `astrald-lab` stage:

* whether the thin prompt reliably triggers the astral-agent skill and the agent
  follows the `node-setup` playbook to a passing `verify.sh`;
* that `objects.store` of a `crypto.private_key` and `tree.set` of the active
  contract succeed under the node's local access without a pre-minted token;
* that the agent keeps `ASTRALD_APPHOST_TOKEN` exported within its own session
  for the User-scoped steps (the skill's "Acting as the User from the CLI"
  section covers this).
