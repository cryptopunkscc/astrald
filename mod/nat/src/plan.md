## 🌐 **Feature: NAT Pair Pool**

### **Scope & Goals**

Implement a local management layer (`PairPool`) responsible for coordinating multiple active **NAT-traversed UDP port pairs** (`pairEntry` instances).
It acts as the intermediary between:

* The **low-level runtime** (`pairEntry` — handles individual pings, sockets, and state)
* The **signalling layer** (which coordinates pair locking and migration between peers)

---
`PairPool` exists entirely **on one node** — it does not communicate over the network directly but provides the local API for the signalling mechanism to work.


* Maintain a registry of all currently active `pairEntry` objects.
* Provide fast lookup by `Nonce` and peer identity.
* Enforce unique pair nonces and avoid duplicates.

### #### 3️⃣ **Lifecycle Control**

### **`Add()` — Register a new NAT-traversed pair**

**Purpose:**
Create and initialize a new `pairEntry` inside the pool after a NAT hole-punch between two peers has succeeded.
Each pair represents a stable, bidirectional UDP path between two public endpoints.

**Flow:**

1. The caller (usually NAT traversal module) supplies a fully formed `EndpointPair` (PeerA + PeerB + Nonce).
2. `PairPool.Add()`:

    * Validates that the nonce is non-zero and unique.
    * Creates a new `pairEntry` object that wraps this `EndpointPair`.
    * Calls `pairEntry.init(localIdentity, isPinger)`:

        * Opens a UDP socket.
        * Starts the internal keepalive mechanism if this side is the pinger.
    * Stores it in the internal `sig.Map`, keyed by `Nonce`.
3. The pair immediately becomes **Idle** and self-maintains its NAT binding through pings/pongs.

**Outcome:**
A new, live `pairEntry` managed by the pool, ready to be used or handed over.
The entry autonomously keeps its UDP mapping alive until it’s either locked, used, or expired.

---

### **`Take()` — Allocate a pair for active use (signalling-coordinated)**

**Purpose:**
Reserve and transfer ownership of an **Idle** pair for use in an upcoming connection or migration — in coordination with the remote peer via the signalling channel.

**High-level idea:**
Both peers maintain the same `pairEntry` (same Nonce) in their respective pools.
When one side wants to use the pair, it performs a coordinated locking handshake over the signalling channel.

---


#### **Key Guarantees**

* **Atomic safety:**
  No two concurrent goroutines can claim the same pair — `CompareAndSwap` protects both `pairEntry` and `PairPool` operations.

* **Symmetric coordination:**
  Each peer maintains its own independent pool, but signalling keeps their state transitions in sync.

* **Strict silence:**
  During `Locked` and `InUse` phases, all UDP traffic ceases, guaranteeing clean handover and no NAT interference.



### TODO

- [x] pairEntry that keep alive traversed UDP socket
- [x] pairPool draft
- [] opPairHandover FSM 
- [] usage of opPairHandover in pair pool
