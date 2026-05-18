# App

An `App` is an external process that uses a Host `Node` over `apphost` IPC.

An App has no direct network presence. The Host owns routing, transport,
encryption, and zone enforcement.

**IPC handshake**

* Host sends `Identity` + alias.
* App may send an `AccessToken`.
* Host returns `GuestID`.
* Missing or expired token gives `Anyone` GuestID.
* `ZoneNetwork` is stripped without a valid token.

**AccessToken**

Time-bounded Host-issued credential that binds a network `Identity` to an IPC
connection.

* Valid token gives an authenticated GuestID.
* Multiple Apps may connect with different GuestIDs.

**Capabilities**

* Send queries. The Host enforces zones, attaches metadata, and bridges the
  Session back.
* Register handlers. The Host forwards inbound queries for the App Identity.
  Handlers disappear on disconnect.

**App record**

`App` (`mod.apphost.app`) is the installed-app record:
`{AppID, HostID, InstalledAt}`.

**AppContract**

`AppContract` is an `auth.SignedContract` with:

* Issuer = AppID.
* Subject = NodeID.
* Permit granting `RelayForAction`.

Uses:

* Relay authorization, attached to outbound queries.
* Relay hints.
* Identity proof.

`OpInstallApp` creates and signs the contract, stores it in the local repo and
auth index, then pushes it to the local swarm.
