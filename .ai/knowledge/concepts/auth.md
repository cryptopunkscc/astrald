# Auth

Authorization answers: "may this action's actor perform this action?"

`Authorize` returns `bool`. Denial does not throw.

Grant rule:

* Any handler returning `true` grants.
* All handlers must deny for denial.

## Actions

* Actions are typed objects embedding `auth.Action` (`Nonce` + `ActorId`).
* Actor identity lives inside the action object, not as a separate argument.
* `ActionObject` interface: `Id() Nonce`, `Actor() *Identity`, `SetActor(*Identity)`, plus `astral.Object`.

## Handlers

* Handlers are registered by action type.
* `TypedHandler` declares its type with `ActionType() string`.
* `Func[T]` adapts a typed callback.
* `Func[T]` type-asserts the action to `T` and returns `false` on mismatch.

## Contract Delegation

If local handlers deny, `Authorize` checks delegation:

* Search active `SignedContract`s where `Subject == actor`.
* Require a permit that covers the action.
* Re-run handlers with actor replaced by the contract's `Issuer`.
* First `true` grants.

## Contracts

`Contract`:

* Shape: `{Issuer, Subject, Permits, ExpiresAt}`.
* Signed by both parties.
* Implements `crypto.SignableTextObject`.
* Signable hash is the contract ObjectID hash.

`Permit`:

* Names an action type.
* May include constraints.

`SignedContract`:

* Wraps `*Contract` with `IssuerSig` and `SubjectSig` (`*crypto.Signature`).
* Both signatures must be present and valid before indexing.

## Constraints

`Constrainable` is an optional interface on actions: `ApplyConstraints(*Bundle) bool`.

Actions that do not implement it always pass the constraint check.

## Built-Ins

`SudoAction`:

* Built-in action type.
* Requests permission for actor to act as `AsID`.
* Grants only when `actor == AsID`.

`ContractQueryBuilder`:

* Fluent query.
* Filters by issuer, subject, or action.
* `Find` reconstructs and returns active signed contracts from the DB.
