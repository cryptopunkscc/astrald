# link-over-tor

Drives the Qwen operator on node1 to **re-establish the swarm link to the peer (node2)
over Tor** after node2 left the LAN, and to confirm the link rides over Tor — following
the astral-agent skill's *linking-over-tor* playbook. The dead LAN link still shows as
up (astrald has no keepalive), so the agent must *force* the Tor link
(`nodes.new_link -strategies tor`) using node2's onion, which `leave-lan` cached on node1
before the cut. The agent records the peer's onion address and the link's transport in
`~/tor.json` (`peer_onion`, `link_network`).

`verify.py` independently confirms node1 holds a link to the peer whose `Network` is
`tor` (`nodes.links`) — it asserts the transport, not the `.onion` endpoint string, since
an inbound tor link legitimately has no remote onion — and cross-checks the agent's
record. Agent-driven. Final task of `tor-link.story`; produces the `two-nodes-tor` stage.
