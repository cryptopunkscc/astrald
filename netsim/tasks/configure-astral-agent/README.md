# configure-astral-agent

A netsim task that installs the `astral-agent` skill into the Qwen Code operator
on a VM, so the operator can drive astrald from the skill's knowledge (the
astral-docs corpus + playbooks) instead of having every procedure spelled out in
each task prompt.

```
configure-astral-agent [--vm <host>] [--user <name>]   # default: node1, tester
```

After it runs, `~<user>/.qwen/skills/astral-agent` exists (SKILL.md with
frontmatter, `references/`, and the `astral-docs` mount). Run standalone against
the lab stage with:

```sh
SATFORGE_SKILLS_DEPLOY_KEY=~/.ssh/satforge_skills_deploy \
  netsim task --stage astrald-lab --save astrald-operator configure-astral-agent
```

## Setup (one-time, on the netsim host)

The host running the sims must own a deploy key for the private repo:

```sh
# 1. generate a keypair (keep the private half on the host)
ssh-keygen -t ed25519 -f ~/.ssh/satforge_skills_deploy -N '' -C netsim-skills-deploy

# 2. register the PUBLIC half on GitHub:
#    satforgedev/skills -> Settings -> Deploy keys -> Add -> paste
#    ~/.ssh/satforge_skills_deploy.pub   (read-only is enough)

# 3. point the env at the PRIVATE key (export it, or prefix each netsim run)
export SATFORGE_SKILLS_DEPLOY_KEY=~/.ssh/satforge_skills_deploy
```

`SATFORGE_SKILLS_DEPLOY_KEY` is a **path to the private deploy-key file**. netsim
runs this task as a subprocess and passes the env through, so exporting it once
covers every `netsim story` / `netsim task` invocation.

## How it works — deploy key, clone in the VM

`satforgedev/skills` is **private**, so the **host** owns the deploy key and the
VM never carries GitHub credentials of its own. `run.sh` (host) reads the private
key from `$SATFORGE_SKILLS_DEPLOY_KEY` and base64-ships it into the VM over one
`netsim ssh` argv. The guest then, as the operator:

1. installs the key at `~/.ssh/skills_deploy` and clones
   `git@github.com:satforgedev/skills` via `GIT_SSH_COMMAND` (parent repo over
   SSH/deploy-key; the `astral-docs` submodule is public HTTPS — no key needed);
2. builds the `satforge-skills` linker (Go is already on the node from
   `install-astrald`);
3. `satforge-skills link astral-agent --target qwen` → installs into
   `~/.qwen/skills/astral-agent` (Qwen Code reads `SKILL.md`, frontmatter intact,
   from there). The clone stays in `~/satforge-skills`, so the install's symlinks
   resolve and the operator can re-link/pull other skills later.

Idempotent: re-running `git pull`s the default branch, `unlink`s, then `link`s again.

## Environment

| Var | Default | Meaning |
|---|---|---|
| `SATFORGE_SKILLS_DEPLOY_KEY` | *(required)* | host path to the private deploy key for the repo |
| `SATFORGE_SKILLS_REPO` | `git@github.com:satforgedev/skills` | repo SSH URL (clones the default branch) |

## Security note

For now the deploy key is **left in the VM** (and therefore in the saved
snapshot) — simplest, and lets the operator re-pull skills. This is a private key
inside a shareable stage; we may switch to wiping it before the snapshot (inject
→ clone/build/link → remove key) if that exposure matters. See the `NOTE` in
`run.sh`.

If outbound SSH:22 is ever blocked in the sim, point `SATFORGE_SKILLS_REPO` at
`ssh://git@ssh.github.com:443/satforgedev/skills`.

## Scope

Installs exactly one skill (`astral-agent`). `node2` is untouched — only the
operator node needs it.
