# install-astrald

A netsim task that builds `astrald` from source and installs it as a systemd
service on target VMs. `run.sh` builds, installs, and enables the unit;
`verify.sh` independently confirms the node answers. The service is left enabled
but stopped, so the netsim stage snapshots cleanly and astrald autostarts when
the stage boots. See [Running astrald as a service](../../../docs/running-as-a-service.md)
for the unit file and operational details.

```
install-astrald [--vm <host>]... [--ref <git-ref>]
```

* No `--vm`: every running VM in the simulation, derived from
  `netsim vm ls --json`.
* `--vm <host>` (repeatable): restrict to the named hosts.
* `--ref <git-ref>`: build a branch or tag via a shallow `--branch` clone instead
  of the default branch.

Each target receives, in one ssh call: `git` and `curl` ensured, Go from the
official tarball, `astrald` and `astral-query` built to `/usr/local/bin`, and a
systemd unit installed and enabled. astrald is started briefly to confirm it
answers `astral-query localnode:.spec`, then stopped for snapshotting.

Use the task in a story (see the [netsim README](../../README.md#lab)), or run it
standalone against an existing stage with
`netsim task --stage <in> --save <out> install-astrald`.

## Execution model

`run.sh` and `verify.sh` run on the host, with the simulation root as the working
directory. They reach each guest with `netsim ssh <host> -- <cmd>` and land as
`root`.

Everything after `--` is one argv element; ssh joins argv with spaces and the
guest shell re-parses it. The whole remote program is sent as a single string:
parameters as an assignment prefix, the body in a single-quoted heredoc
(`<<'EOS'`) so host-side `$...` reach the guest unexpanded.

```sh
netsim ssh "$vm" -- "repo='$REPO' ref='$REF' go_ver='$GO_VERSION'; $REMOTE_BODY"
```

## Build and run facts

* Go is installed from the official tarball. astrald's `go.mod` requires
  `go >= 1.25.0`; the apt package is older. The download is arch-aware
  (`x86_64`→`amd64`, `aarch64`→`arm64`).
* The clone is `git clone --depth 1` over HTTPS, never `--recursive`. The only
  submodule (`.ai/system`) is an SSH-only docs repo and is not needed for the
  build.
* The build sets `CGO_ENABLED=0`. astrald uses pure-Go SQLite and needs no C
  toolchain.
* Build targets carry the `./` prefix: `go build -o /usr/local/bin/astrald
  ./cmd/astrald`, and the same for `./cmd/astral-query`. `go build cmd/astrald`
  fails; `go build .` at the repo root builds a do-nothing stub.
* The service runs `astrald -root /var/lib/astrald` with `Environment=HOME=/root`.
  Default config and data paths derive from `$HOME`, which systemd does not set.
  The unit sets `KillSignal=SIGINT` so `systemctl stop` shuts astrald down
  gracefully (astrald traps SIGINT, not SIGTERM).
* First start auto-generates the node identity, a `secp256k1` key at
  `/var/lib/astrald/config/node_key`, with no prompt and no TTY.
* The liveness probe is `astral-query localnode:.spec`. The op is built-in and
  always available; it streams the node's operation spec over the local apphost
  API (`tcp:127.0.0.1:8625`, anonymous access by default). Exit code 0 means
  healthy.
* apt calls pass `-o DPkg::Lock::Timeout=120`; `cloud-init` can hold the dpkg
  lock on a fresh boot. Readiness is never gated on `ping`; guest ICMP is
  disabled.
