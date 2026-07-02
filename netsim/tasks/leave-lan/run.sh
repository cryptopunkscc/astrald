#!/bin/sh
# leave-lan: make <vm> (node2, the node that "leaves") genuinely leave the 10.77 LAN, so
# astrald's tor module + the swarm link maintainer re-link to <peer> (node1) over Tor.
#
# Two steps, both on the host:
#   1. Seed <peer> with <vm>'s onion WHILE THE LAN IS STILL UP — once the LAN is gone the
#      peer can no longer ask <vm> for its address, so it needs the .onion cached first.
#   2. Withdraw <vm>'s own 10.77 LAN address (ip addr flush). astrald has no carrier/
#      operstate monitor: it polls net.InterfaceAddrs() every 3s and advertises one tcp
#      endpoint per assigned IP, so removing the address is what it observes as "left the
#      network" — it drops the 10.77 endpoint and re-links over Tor. (A packet-filter DROP,
#      or even a link/carrier down, leaves the IPv4 address in place and is invisible to
#      that monitor.) SSH/management rides the separate WAN NIC and is untouched.
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

# 2) make <vm> leave the LAN: withdraw its own 10.77 address (and drop the NIC for realism).
#    Removing the address takes its connected /24 route with it, so <vm> has no address on
#    and no route to the LAN — it has genuinely left at the IP layer, which is exactly what
#    astrald observes (see the header). No peer IP needed: <vm> drops its own membership.
CUT_BODY=$(cat <<'EOS'
set -eu
# the NIC holding the 10.77 LAN address is nic2; SSH rides the separate WAN NIC, untouched.
lan_if=$(ip -o -4 addr show | awk '$4 ~ /^10\.77\./ {print $2; exit}')
[ -n "$lan_if" ] || { echo "leave-lan: no 10.77 LAN interface on $(hostname)" >&2; exit 1; }
lan_ip=$(ip -o -4 addr show dev "$lan_if" | awk '$4 ~ /^10\.77\./ {print $4; exit}')
ip addr flush dev "$lan_if"   # RTM_DELADDR: drops the address AND its connected /24 route
ip link set "$lan_if" down    # carrier/admin down too, so the NIC is faithfully "gone"
echo "leave-lan: $(hostname) withdrew $lan_ip from $lan_if (left the LAN)"
EOS
)
echo "leave-lan: $VM leaving the LAN (withdrawing its 10.77 address) ..."
# shellcheck disable=SC2029
netsim ssh "$VM" -- "$CUT_BODY"
echo "leave-lan: done on $VM"
