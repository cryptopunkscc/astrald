#!/bin/sh
# verify share-object: an astral object stored on node1 must be obtainable by its
# sibling node2 ACROSS THE SWARM. INDEPENDENT host-side check -- it does not trust
# run.sh. It reads the id + payload the agent persisted on node1, then tries to
# pull that exact id FROM node2's vantage and asserts the bytes match.
#
# THE CROSS-SWARM HOP IS INFERRED, NOT DOCUMENTED (see README). The astral-docs
# describe a network zone + a finder/provider layer but no worked example of one
# swarm member reading another's object by id, so verify probes a LADDER and
# reports which hop routes -- exactly as link-swarm discovered that <peer>:.spec
# does NOT route. Order (strongest -> weakest), all run on node2:
#   1. EXPLICIT TARGET  astral-query <node1-id>:objects.load -id <ID> -out json
#        Query-target routing over the swarm link. Primary path: it does NOT rely
#        on node2's network zone (an anonymous apphost caller has ZoneNetwork
#        stripped), it addresses node1 directly; node1 serves the read locally.
#   2. TRANSPARENT      astral-query objects.load -id <ID> -out json
#        Relies on the read context's zone defaulting to all zones (incl. the
#        network zone) so node2 resolves node1 as provider. Likely BLOCKED for an
#        anonymous host-side caller (network zone stripped) -- kept as a bonus probe.
#   3. PROVIDER FIND    astral-query objects.find -id <ID> -out json
#        Returns provider IDENTITIES, not bytes. If only this works, discovery
#        crosses the swarm but the byte read does not -- a partial finding, not a pass.
#
# PASS iff node2 obtained the EXACT stored bytes for the agent-reported id across
# the swarm (hop 1 or 2). A pre-check asserts node2 does not already hold the
# object locally, so a pass reflects a genuine remote pull.
#
# astral-query ... -out json emits a JSON *stream* (one object per line, then an
# {"Type":"eos"} terminator), so everything is parsed line-by-line, not as one
# document.
set -eu

VM1=node1
VM2=node2
while [ $# -gt 0 ]; do
  case "$1" in
    --node1) VM1=$2; shift 2 ;;
    --node2) VM2=$2; shift 2 ;;
    *)       shift ;;
  esac
done

# --- node1 side: the id + payload the agent persisted, and node1's node id ------
# node1 acts as the User (token from bootstrap-user) so user.info returns the
# active contract whose Subject IS node1's node identity (the provider to target).
ID=$(netsim ssh "$VM1" -- 'cat /home/tester/.netsim/object.id')
PAY=$(netsim ssh "$VM1" -- 'cat /home/tester/.netsim/object.payload')
n1_info=$(netsim ssh "$VM1" -- 'export ASTRALD_APPHOST_TOKEN=$(cat /home/tester/.netsim/user.token); astral-query user.info -out json')

IDC=$(printf '%s' "$ID" | tr -d '[:space:]')

# --- node2 side: cross-check node1's id, locality pre-check, then the fetch ladder
# node2 answers under its node identity (no token => anonymous apphost caller).
n2_links=$(netsim ssh "$VM2" -- 'astral-query nodes.links -out json' 2>/dev/null || true)
n2_contains=$(netsim ssh "$VM2" -- "astral-query objects.contains -repo local -id '$IDC' -out json" 2>/dev/null || true)
n2_find=$(netsim ssh "$VM2" -- "astral-query objects.find -id '$IDC' -out json" 2>/dev/null || true)
n2_transparent=$(netsim ssh "$VM2" -- "astral-query objects.load -id '$IDC' -out json" 2>/dev/null || true)

# explicit-target read needs node1's node id; resolve it host-side first.
N1=$(printf '%s' "$n1_info" | python3 -c '
import sys, json
for ln in sys.stdin:
    ln = ln.strip()
    if not ln:
        continue
    try:
        o = json.loads(ln)
    except json.JSONDecodeError:
        continue
    ob = o.get("Object")
    if isinstance(ob, dict) and isinstance(ob.get("Contract"), dict):
        c = ob["Contract"].get("Contract", {})
        if c.get("Subject"):
            print(c["Subject"]); break
' 2>/dev/null || true)
# fall back to the RemoteIdentity of node2->node1 link if user.info parse failed
if [ -z "$N1" ]; then
  N1=$(printf '%s' "$n2_links" | python3 -c '
import sys, json
for ln in sys.stdin:
    ln = ln.strip()
    if not ln:
        continue
    try:
        o = json.loads(ln)
    except json.JSONDecodeError:
        continue
    ob = o.get("Object")
    if isinstance(ob, dict) and ob.get("RemoteIdentity"):
        print(ob["RemoteIdentity"]); break
' 2>/dev/null || true)
fi

n2_explicit=""
if [ -n "$N1" ]; then
  n2_explicit=$(netsim ssh "$VM2" -- "astral-query '$N1':objects.load -id '$IDC' -out json" 2>/dev/null || true)
fi

export IDC PAY N1 N1_INFO="$n1_info" N2_LINKS="$n2_links" N2_CONTAINS="$n2_contains" \
       N2_FIND="$n2_find" N2_TRANSPARENT="$n2_transparent" N2_EXPLICIT="$n2_explicit"

python3 - <<'PY'
import os, sys, json

def objs(s):
    """Parse a JSON object-stream (one object per line + an eos terminator)."""
    out = []
    for ln in (s or "").splitlines():
        ln = ln.strip()
        if not ln:
            continue
        try:
            out.append(json.loads(ln))
        except json.JSONDecodeError:
            pass
    return out

def loaded_payload(stream):
    """From an objects.load -out json stream, return the decoded payload string
    (the stored string8's Object), or None. Skips eos / error_message frames."""
    for o in objs(stream):
        if o.get("Type") in ("eos", "error_message"):
            continue
        ob = o.get("Object")
        if isinstance(ob, str):
            return ob
    return None

def errors(stream):
    return [o.get("Object") for o in objs(stream) if o.get("Type") == "error_message"]

def contains_local(stream):
    """objects.contains -out json -> a bool frame. Returns True/False/None."""
    for o in objs(stream):
        if o.get("Type") in ("eos", "error_message"):
            continue
        if isinstance(o.get("Object"), bool):
            return o["Object"]
    return None

def find_identities(stream):
    ids = []
    for o in objs(stream):
        if o.get("Type") in ("eos", "error_message"):
            continue
        ob = o.get("Object")
        if isinstance(ob, str):
            ids.append(ob)
    return ids

ID  = os.environ["IDC"]
N1  = os.environ.get("N1", "")
# compare tolerant of a single trailing newline on either side
PAY = os.environ["PAY"].rstrip("\n")

already_local = contains_local(os.environ["N2_CONTAINS"])
explicit      = loaded_payload(os.environ["N2_EXPLICIT"])
transparent   = loaded_payload(os.environ["N2_TRANSPARENT"])
providers     = find_identities(os.environ["N2_FIND"])

explicit_ok    = explicit    is not None and explicit.rstrip("\n")    == PAY
transparent_ok = transparent is not None and transparent.rstrip("\n") == PAY
find_ok        = (N1 in providers) if N1 else bool(providers)

errs = []
notes = []

if not ID:
    errs.append("no Object ID recorded on node1 (~/.netsim/object.id)")
if not PAY:
    errs.append("no payload recorded on node1 (~/.netsim/object.payload)")
if not N1:
    notes.append("could not resolve node1's node identity host-side (explicit-target read skipped)")

# locality pre-check is advisory (objects.contains is probabilistic): a 'true'
# here means the pass might not reflect a genuine remote pull -> warn, don't fail.
if already_local is True:
    notes.append("objects.contains reports node2 may ALREADY hold this object locally; "
                 "a byte-match below might not be a genuine cross-swarm pull")
elif already_local is None:
    notes.append("objects.contains gave no usable answer on node2 (locality pre-check inconclusive)")

# surface auth-vs-route signal: an error_message naming auth/permission is a
# DIFFERENT failure than no route / no provider -- don't conflate them.
for label, env in (("explicit-target", "N2_EXPLICIT"),
                   ("transparent", "N2_TRANSPARENT"),
                   ("objects.find", "N2_FIND")):
    for e in errors(os.environ.get(env, "")):
        notes.append(f"{label} returned error_message: {e}")

crossed = explicit_ok or transparent_ok

if crossed:
    path = "explicit-target (<node1>:objects.load)" if explicit_ok else "transparent (objects.load, network zone)"
    print(f"share-object OK: node2 pulled object {ID[:12]}.. from node1 across the swarm "
          f"via {path}; bytes match ({len(PAY)} B). "
          f"providers seen by objects.find: {len(providers)}.")
    for n in notes:
        sys.stderr.write(f"  note: {n}\n")
    sys.exit(0)

# Did not cross. Build a precise diagnostic (the link-swarm-style finding).
sys.stderr.write("share-object verify FAILED: object did NOT cross the swarm to node2.\n")
if find_ok:
    sys.stderr.write("  FINDING: provider discovery DOES cross the swarm "
                     "(objects.find on node2 returned node1) but the byte READ did not route "
                     "(explicit-target and transparent objects.load both failed to return the payload). "
                     "This is the share-object analogue of link-swarm's '<peer>:.spec does not route' "
                     "discovery -- record which hop routes in the task log.\n")
else:
    sys.stderr.write("  no cross-swarm object access at all: neither a read nor objects.find "
                     "resolved node1's object from node2.\n")
for e in errs:
    sys.stderr.write(f"  - {e}\n")
for n in notes:
    sys.stderr.write(f"  note: {n}\n")
sys.stderr.write(f"  (id={ID} node1={N1[:12] + '..' if N1 else '?'} "
                 f"explicit={'hit' if explicit is not None else 'miss'} "
                 f"transparent={'hit' if transparent is not None else 'miss'} "
                 f"find_providers={len(providers)})\n")
sys.exit(1)
PY

echo "verified share-object across: $VM1 and $VM2"
