# netsim scenarios for astrald

Test scaffolding that drives `netsim` to build and run `astrald` on a simulated
LAN. It contains no astrald Go source and modifies none.

`netsim` boots Ubuntu 26.04 cloud-image VMs on `10.77.0.0/24` with per-VM NAT. A
*task* is a host-side script that configures the VMs. A *story* runs a list of
tasks in one simulation and saves a named *stage*. `lab.story` builds the
`astrald-lab` stage: two nodes running astrald, with a Qwen Code operator on
`node1`.

## Layout

```
netsim/
  tasks/
    install-astrald/               # build + run astrald as a service (tasks/install-astrald/README.md)
    configure-astral-agent/        # install the astral-agent skill into the qwen operator
      run.sh / verify.sh / README.md   # each task: installs on target VMs + independent re-check
  lab.story                        # full lab in one simulation -> stage astrald-lab
  link.sh                          # register tasks with netsim (idempotent; re-run anytime)
  README.md
```

## Registering tasks

`netsim` discovers tasks only under `~/.local/share/netsim/tasks/`. `link.sh`
symlinks every task under `tasks/` — each folder containing a `run.sh` — there.
It is idempotent; re-run it after adding a task. The symlinks leave netsim's
shipped builtins intact.

```sh
./netsim/link.sh
netsim tasks        # confirm: install-astrald is listed as a user task
```

## Lab

`lab.story` builds the full lab in one simulation: two nodes running astrald and
a Qwen Code operator on `node1`, equipped with the `astral-agent` skill.

```
# lab.story — the astrald lab, built in one netsim simulation.
# Result: a single stage with two nodes running astrald and a Qwen Code
# operator on node1, equipped with the astral-agent skill.
add-vm --hostname node1
add-vm --hostname node2
install-astrald
install-qwen-code --vm node1 --create-user
configure-astral-agent --vm node1
```

A story is a plain-text file with one `task [args...]` per line, shell-style
quoting, and `#` for full-line or trailing comments. `netsim story` boots one
simulation, runs the listed tasks in order in the same VMs, and saves a single
stage at the end. It stops at the first failing task. Order is significant:

* `add-vm --hostname node1` and `add-vm --hostname node2` use the `add-vm`
  builtin; they create the two plain Ubuntu VMs on the LAN.
* `install-astrald` is the [custom task](tasks/install-astrald/README.md); with no
  `--vm` it installs astrald on every running VM, so on both nodes. It runs
  `run.sh` then `verify.sh` and fails the story unless astrald builds, starts, and
  answers `astral-query localnode:.spec` on every node. The service is left
  enabled and running, so the stage snapshots a live node that resumes
  already-running on restore.
* `install-qwen-code --vm node1 --create-user` uses the `install-qwen-code`
  builtin; it installs the Qwen Code CLI on `node1` and points it at the
  inference endpoint. The builtin installs for user `tester`, which does not
  exist on a fresh cloud image, so `--create-user` is required. `node2` stays a
  plain astrald peer.
* `configure-astral-agent --vm node1` is a [custom task](tasks/configure-astral-agent/README.md);
  it installs the `astral-agent` skill into the Qwen Code operator so it can drive
  astrald from the skill's knowledge. The host must have `SATFORGE_SKILLS_DEPLOY_KEY`
  set (a deploy key for the private skills repo) — see its README.

Both VMs must exist and run before `install-astrald`, astrald must be present
before the Qwen Code operator is layered on `node1`, and the operator must exist
before its skill is configured.

Register the custom tasks once (see [Registering tasks](#registering-tasks)),
then build the lab:

```sh
./netsim/link.sh
export SATFORGE_SKILLS_DEPLOY_KEY=~/.ssh/satforge_skills_deploy   # see tasks/configure-astral-agent
netsim story --stage null --save astrald-lab netsim/lab.story
```

The result is the stage `astrald-lab`: `node1` and `node2` running astrald, with a
Qwen Code operator on `node1` equipped with the `astral-agent` skill. Re-enter it
with `netsim shell --stage astrald-lab`.

## Scope

v1 installs and runs astrald on each node as two independent nodes. Linking the
nodes and verifying a live session is a later phase.

Fresh nodes broadcast on UDP 8822 through the `ether` and `nearby` modules and
discover each other on a shared L2 LAN. v1 asserts nothing about discovery.
