"""Shared host-side verify library for the netsim astral scenarios.

Each task's verify.py reaches this through a realpath shim that crosses netsim's
per-task symlink:

    import os, sys
    sys.path.insert(0, os.path.join(
        os.path.dirname(os.path.dirname(os.path.realpath(__file__))), "_lib"))
    import netsim_astral as na

It centralises the two halves every verifier shares:

  * transport -- ssh()/file readers/all_running_vms()/peer_lan_ip(): unchanged
    subprocess plumbing for reading the agent's recorded artifacts and probing
    inside a VM.

  * queries -- connect(vm, token=...) yields a Node whose .call(op, ...) returns
    a list of astral.AstralObject. Queries go through the astral-py typed client
    (reached host-side over an ssh -L forward of the VM's WebSocket apphost
    port), falling back to the lockstep Go `astral-query` CLI -- same JSON,
    parsed with astral-py's from_json_envelope -- whenever the client can't
    serve an op (pinned in SHELL_OPS, or it raised). Both paths return the same
    list[AstralObject], so the interrogators below are transport-agnostic.

astral-py is imported from an editable checkout (no pip needed on this host):
$ASTRALPY_SRC, else ~/work/satforge/astral-py/master/src.
"""
import contextlib
import json
import os
import socket
import subprocess
import sys
import time

# --- astral-py (editable checkout; pip-free) ---------------------------------
_ASTRALPY_SRC = os.environ.get("ASTRALPY_SRC") or os.path.expanduser(
    "~/work/satforge/astral-py/master/src")
if os.path.isdir(_ASTRALPY_SRC) and _ASTRALPY_SRC not in sys.path:
    sys.path.insert(0, _ASTRALPY_SRC)
import astral  # noqa: E402
from astral.encoding import from_json_envelope  # noqa: E402

# apphost WebSocket port inside each VM (binds 0.0.0.0; reachable via ssh -L).
WS_PORT = 8624

# Ops to keep on the Go astral-query CLI instead of the astral-py client.
# Populated by the smoke-test triage when the client disagrees with the CLI on a
# specific op (a silent mismatch the auto-fallback can't catch). Empty => every
# op tries the client first.
SHELL_OPS = set()


# --- transport: subprocess into the VM ---------------------------------------
def ssh(vm, remote):
    """Run `netsim ssh <vm> -- <remote>` on the host; return stdout (best-effort)."""
    p = subprocess.run(["netsim", "ssh", vm, "--", remote],
                       capture_output=True, text=True)
    return p.stdout


def read_file(vm, path):
    """Contents of <path> on the VM, trailing newline stripped ("" on error)."""
    return (ssh(vm, f"cat {path}") or "").rstrip("\n")


def read_json(vm, path):
    """<path> on the VM parsed as a dict ({} on error)."""
    try:
        return json.loads(ssh(vm, f"cat {path}") or "{}") or {}
    except json.JSONDecodeError:
        return {}


def home_json(vm, name):
    """An agent artifact under /home/tester/<name>, parsed as a dict."""
    return read_json(vm, f"/home/tester/{name}")


def all_running_vms():
    """Hostnames of the running VMs in the current simulation."""
    out = subprocess.run(["netsim", "vm", "ls", "--json"],
                         capture_output=True, text=True).stdout
    try:
        return [v["hostname"] for v in json.loads(out or "[]")
                if v.get("state") == "running"]
    except json.JSONDecodeError:
        return []


def peer_lan_ip(peer):
    """The 10.77.* LAN address of <peer> ("" if none)."""
    for tok in (ssh(peer, "hostname -I") or "").split():
        if tok.startswith("10.77."):
            return tok
    return ""


# --- queries: astral-py client over an ssh -L forward, Go-CLI fallback -------
def parse_cli(raw):
    """Parse `astral-query -out json` output into AstralObjects (eos dropped)."""
    out = []
    for ln in (raw or "").splitlines():
        ln = ln.strip()
        if not ln:
            continue
        try:
            obj = from_json_envelope(json.loads(ln))
        except Exception:
            continue
        if not obj.is_eos:
            out.append(obj)
    return out


def _free_port():
    s = socket.socket()
    s.bind(("127.0.0.1", 0))
    port = s.getsockname()[1]
    s.close()
    return port


def _wait_port(port, timeout=10.0):
    deadline = time.time() + timeout
    while time.time() < deadline:
        try:
            with socket.create_connection(("127.0.0.1", port), timeout=0.5):
                return True
        except OSError:
            time.sleep(0.1)
    return False


class Node:
    """A handle to one VM's apphost: .call(op, ...) -> list[AstralObject]."""

    def __init__(self, vm, client, token):
        self.vm = vm
        self._client = client
        self._token = token

    @property
    def uses_client(self):
        return self._client is not None

    def _via_shell(self, op, args, target):
        q = f"{target}:{op}" if target else op
        flags = "".join(f" -{k} '{v}'" for k, v in (args or {}).items())
        tok = f"export ASTRALD_APPHOST_TOKEN={self._token}; " if self._token else ""
        return parse_cli(ssh(self.vm, f"{tok}astral-query {q}{flags} -out json"))

    def call(self, op, args=None, target=None):
        """Run an apphost op; return its result objects (eos dropped, errors kept).

        Routes through the astral-py client unless the op is pinned in SHELL_OPS
        or no client is available; on any client error, falls back to the Go CLI.
        """
        base = op.split(":")[-1]
        if self._client is None or op in SHELL_OPS or base in SHELL_OPS:
            return self._via_shell(op, args, target)
        try:
            with self._client.query(op, args or None, target=target) as st:
                return list(st)
        except Exception:
            return self._via_shell(op, args, target)


@contextlib.contextmanager
def connect(vm, token=None):
    """Yield a Node for <vm>.

    Opens an ssh -L forward of the VM's WebSocket apphost port (using netsim's
    own $NETSIM_SSH_CONFIG) and an astral-py client over it. If the forward or
    client can't be established, yields a shell-only Node so verification still
    runs via the Go CLI.
    """
    cfg = os.environ.get("NETSIM_SSH_CONFIG")
    client = None
    tunnel = None
    if cfg:
        try:
            port = _free_port()
            tunnel = subprocess.Popen(
                ["ssh", "-F", cfg, "-o", "ExitOnForwardFailure=yes",
                 "-L", f"{port}:127.0.0.1:{WS_PORT}", "-N", vm],
                stdout=subprocess.DEVNULL, stderr=subprocess.DEVNULL)
            if _wait_port(port):
                client = astral.connect(f"ws://127.0.0.1:{port}/.ws", token=token)
        except Exception:
            client = None
    try:
        yield Node(vm, client, token)
    finally:
        try:
            if client is not None:
                client.close()
        except Exception:
            pass
        if tunnel is not None:
            tunnel.terminate()
            try:
                tunnel.wait(timeout=3)
            except Exception:
                tunnel.kill()


# --- interrogators: list[AstralObject] -> extracted value --------------------
def _values(objs):
    return [o.value for o in objs if not o.is_eos]


def contract(objs):
    """(Issuer, Subject) of the active contract from a user.info result."""
    for v in _values(objs):
        if isinstance(v, dict) and isinstance(v.get("Contract"), dict):
            c = v["Contract"].get("Contract", {})
            return c.get("Issuer"), c.get("Subject")
    return None, None


def linked_sibling(objs):
    """Identity of the first Linked sibling in a user.swarm_status result."""
    for v in _values(objs):
        if isinstance(v, dict) and v.get("Linked"):
            return v.get("Identity")
    return None


def swarm_identities(objs):
    """Set of node identities in a user.swarm_status result."""
    ids = set()
    for v in _values(objs):
        if isinstance(v, dict) and v.get("Identity"):
            ids.add(v["Identity"])
    return ids


def has_link_to(objs, ident):
    """True if a nodes.links result holds an active link to <ident>."""
    return any(isinstance(v, dict) and v.get("RemoteIdentity") == ident
               for v in _values(objs))


def _contains_identity(value, ident):
    if isinstance(value, str):
        return value == ident
    if isinstance(value, dict):
        return any(_contains_identity(v, ident) for v in value.values())
    if isinstance(value, list):
        return any(_contains_identity(v, ident) for v in value)
    return False


def is_expelled(objs, ident):
    """True if a user.list_expelled result bans <ident> (nested Subject match)."""
    return any(_contains_identity(o.value, ident) for o in objs
               if o.type not in ("eos", "error_message"))


def loaded_payload(objs):
    """The decoded string payload from an objects.load result, or None."""
    for o in objs:
        if o.type in ("eos", "error_message"):
            continue
        if isinstance(o.value, str):
            return o.value
    return None


def error_messages(objs):
    """The error_message strings in a result stream."""
    return [o.value for o in objs if o.type == "error_message"]


def endpoint_addr(ep):
    """Address string of an exonet.Endpoint (bare or {Type,Object})."""
    if isinstance(ep, str):
        return ep
    if isinstance(ep, dict):
        o = ep.get("Object")
        return o if isinstance(o, str) else ""
    return ""


def tor_links(objs):
    """(RemoteIdentity, endpoint-address) for links whose Network == 'tor'."""
    out = []
    for v in _values(objs):
        if isinstance(v, dict) and str(v.get("Network")) == "tor":
            out.append((str(v.get("RemoteIdentity", "")),
                        endpoint_addr(v.get("RemoteEndpoint"))))
    return out


def resolve_onion(objs):
    """The .onion address from a nodes.resolve_endpoints result, or None."""
    for v in _values(objs):
        if isinstance(v, dict):
            a = endpoint_addr(v.get("Endpoint"))
            if ".onion" in a:
                return a
    return None
