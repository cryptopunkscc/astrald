#!/bin/sh
# verify link-swarm: node1 and node2 must be linked into one User swarm.
# INDEPENDENT both-ends check -- it does not trust run.sh. It pulls raw JSON from
# both nodes and asserts THREE facts on the host; together they prove the swarm
# from both ends:
#   1. both nodes hold an active contract issued by the SAME User
#      (user.info: Issuer == the bootstrap User on each; Subject == that node);
#   2. node1, acting as the User, lists node2 as a Linked sibling
#      (user.swarm_status);
#   3. a mutual authenticated link exists (node2 nodes.links -> node1).
#
# NOTE on "routed query": an earlier plan probed `<peer>:.spec` as the proof.
# That is NOT valid -- node introspection ops (.spec/.id/.ping) are served
# locally and do not route to a sibling by node-id, so they fail even on a fully
# formed swarm. The contract + link + sibling triple above is the real proof.
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

# Pull the User id and the four JSON blobs. Single-quoted remote args so the
# command substitutions run on the guest. node1 acts as the User (token from
# bootstrap-user); node2 answers under its node identity (it holds the contract).
U=$(netsim ssh "$VM1" -- 'cat /home/tester/.netsim/user.id')
n1_info=$(netsim ssh "$VM1" -- 'export ASTRALD_APPHOST_TOKEN=$(cat /home/tester/.netsim/user.token); astral-query user.info -out json')
n1_swarm=$(netsim ssh "$VM1" -- 'export ASTRALD_APPHOST_TOKEN=$(cat /home/tester/.netsim/user.token); astral-query user.swarm_status -out json')
n2_info=$(netsim ssh "$VM2" -- 'astral-query user.info -out json')
n2_links=$(netsim ssh "$VM2" -- 'astral-query nodes.links -out json')

UU=$(printf '%s' "$U" | tr -d '[:space:]')
export UU N1_INFO="$n1_info" N1_SWARM="$n1_swarm" N2_INFO="$n2_info" N2_LINKS="$n2_links"

python3 - <<'PY'
import os, sys, json

def objs(s):
    out = []
    for ln in s.splitlines():
        ln = ln.strip()
        if not ln:
            continue
        try:
            out.append(json.loads(ln))
        except json.JSONDecodeError:
            pass
    return out

def contract(info):
    for o in objs(info):
        ob = o.get("Object")
        if isinstance(ob, dict) and isinstance(ob.get("Contract"), dict):
            c = ob["Contract"].get("Contract", {})
            return c.get("Issuer"), c.get("Subject")
    return None, None

U = os.environ["UU"]
i1, s1 = contract(os.environ["N1_INFO"])
i2, s2 = contract(os.environ["N2_INFO"])

sib = None
for o in objs(os.environ["N1_SWARM"]):
    ob = o.get("Object")
    if isinstance(ob, dict) and ob.get("Linked"):
        sib = ob.get("Identity")
        break

linkback = False
for o in objs(os.environ["N2_LINKS"]):
    ob = o.get("Object")
    if isinstance(ob, dict) and ob.get("RemoteIdentity") == s1:
        linkback = True
        break

errs = []
if not U:               errs.append("no User id recorded on node1 (~/.netsim/user.id)")
if i1 != U:             errs.append(f"node1 contract issuer {i1} != User {U}")
if i2 != U:             errs.append(f"node2 contract issuer {i2} != User {U} (node2 not claimed under this User)")
if not s1:              errs.append("node1 has no active contract subject")
if not s2:              errs.append("node2 has no active contract subject")
if s2 and sib != s2:    errs.append(f"node1's linked sibling {sib} != node2 {s2}")
if not linkback:        errs.append(f"node2 has no active link back to node1 ({s1})")

if errs:
    sys.stderr.write("link-swarm verify FAILED:\n")
    for e in errs:
        sys.stderr.write(f"  - {e}\n")
    sys.exit(1)

print(f"swarm OK: User {U[:8]}.. ; node1 {s1[:8]}.. <-link-> node2 {s2[:8]}.. ; "
      f"both under one User; node1 lists node2 as Linked sibling")
PY

echo "verified swarm link on: $VM1 and $VM2"
