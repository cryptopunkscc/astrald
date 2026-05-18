# Query

A `Query` requests a bidirectional `Session` with a named service on a target
`Identity`. It is the base communication operation.

## Routing

Node tries registered routers in priority order.

Router outcomes:

* Accept: open the `Session`.
* Reject: stop routing.
* Not found: try the next router.

See `rules.md` for return conventions.

## Preprocessors

Preprocessors run before routing, in registration order. They may:

* Attach metadata.
* Add `relay_via` hints.
* Block the query.

They are discovered during `Inject`. apphost uses a preprocessor to attach
AppContracts.

## Gateway

Gateway relays for Nodes that are unreachable directly, such as behind NAT or a
firewall. The application sees a normal `Session`.
