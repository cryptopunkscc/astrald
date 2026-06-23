#!/bin/sh
# leave-lan: sever the LAN path between <vm> (node2, the node that "leaves") and <peer>
# (node1). astrald's tor module + the swarm link maintainer will then re-link over Tor.
#
# Two steps, both on the host:
#   1. Seed <peer> with <vm>'s onion WHILE THE LAN IS STILL UP — once the LAN is gone the
#      peer can no longer ask <vm> for its address, so it needs the .onion cached first.
#   2. nftables-drop all traffic between them on the LAN. The NIC stays up and Internet
#      egress (the WAN NAT, used for Tor) is untouched — only the direct LAN path is cut.
#   leave-lan [--vm <host>] [--peer <host>]    (default: node2 leaves, peer node1)
#
# Both nodes must have Tor up (enable-tor) and the alias <vm> must resolve on <peer>
# (adopt-node). astral-query ops here (resolve_endpoints / add_endpoint) are ungated.
set -eu

VM="node2"; PEER="node1"
while [ $# -gt 0 ]; do
  case "$1" in
    --vm)   [ $# -ge 2 ] || { echo "need host after --vm" >&2; exit 64; }; VM=$2; shift 2 ;;
    --peer) [ $# -ge 2 ] || { echo "need host after --peer" >&2; exit 64; }; PEER=$2; shift 2 ;;
    *)      echo "usage: leave-lan [--vm <host>] [--peer <host>]" >&2; exit 64 ;;
  esac
done

# 1) seed <peer> with <vm>'s onion before the LAN goes away
SEED_BODY=$(cat <<'EOS'
set -eu
torof() {  # read a .onion endpoint address from a resolve_endpoints json stream on stdin
  python3 -c '
import json,sys
def addr(ep):
    if isinstance(ep,str): return ep
    if isinstance(ep,dict):
        o=ep.get("Object"); return o if isinstance(o,str) else ""
    return ""
for ln in sys.stdin:
    ln=ln.strip()
    if not ln: continue
    try: o=json.loads(ln)
    except Exception: continue
    a=addr((o.get("Object") or {}).get("Endpoint"))
    if ".onion" in a: print(a); break'
}
# prefer the local cache (auto-synced over the live link); else ask the leaver directly
onion=$(astral-query nodes.resolve_endpoints -id "$leaver" -out json 2>/dev/null | torof || true)
[ -n "$onion" ] || onion=$(astral-query "$leaver":nodes.resolve_endpoints -id "$leaver" -out json 2>/dev/null | torof || true)
[ -n "$onion" ] || { echo "leave-lan: $(hostname) could not learn $leaver's onion before the cut" >&2; exit 1; }
astral-query nodes.add_endpoint -id "$leaver" -endpoint "tor:$onion" >/dev/null 2>&1 || true
echo "leave-lan: $(hostname) seeded $leaver onion=$onion"
EOS
)
echo "leave-lan: seeding $PEER with $VM's onion ..."
# shellcheck disable=SC2029
netsim ssh "$PEER" -- "leaver='$VM'; $SEED_BODY"

# 2) resolve <peer>'s LAN address and drop it on <vm>
peer_ip=$(netsim ssh "$PEER" -- "hostname -I" | tr ' ' '\n' | grep '^10\.77\.' | head -1)
[ -n "$peer_ip" ] || { echo "leave-lan: could not find $PEER's 10.77 LAN address" >&2; exit 1; }

CUT_BODY=$(cat <<'EOS'
set -eu
export DEBIAN_FRONTEND=noninteractive
command -v nft >/dev/null 2>&1 || {
  apt-get -qq -o DPkg::Lock::Timeout=120 update
  apt-get -qq -y -o DPkg::Lock::Timeout=120 install nftables >/dev/null
}
# A dedicated table so the cut is self-contained and easy to reason about. Chains are
# named netout/netin (not the nft scanner keywords in/out). Flush before adding so a
# re-run yields exactly one rule per direction.
nft add table ip netsimcut 2>/dev/null || true
nft 'add chain ip netsimcut netout { type filter hook output priority 0 ; }' 2>/dev/null || true
nft 'add chain ip netsimcut netin  { type filter hook input  priority 0 ; }' 2>/dev/null || true
nft flush chain ip netsimcut netout 2>/dev/null || true
nft flush chain ip netsimcut netin  2>/dev/null || true
nft add rule ip netsimcut netout ip daddr "$peer_ip" drop
nft add rule ip netsimcut netin  ip saddr "$peer_ip" drop
echo "leave-lan: $(hostname) dropped LAN traffic to/from $peer_ip"
EOS
)
echo "leave-lan: severing LAN path $VM <-> $PEER ($peer_ip) ..."
# shellcheck disable=SC2029
netsim ssh "$VM" -- "peer_ip='$peer_ip'; $CUT_BODY"
echo "leave-lan: done on $VM"
