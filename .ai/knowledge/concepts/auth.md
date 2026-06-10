# Auth

Authorization answers: "may this action's actor perform this action?"

`Authorize` returns `bool`. Denial does not throw.

Grant rule:

* Any handler returning `true` grants.
* All handlers must deny for denial.

## Actions

* Actions are typed objects embedding `auth.Action` (`Nonce` + `ActorId`).
* The actor lives inside the action object, not as a separate argument: `Actor() *Identity`, `SetActor(*Identity)`.
* `ActionObject` interface: `astral.Object`, `Id() Nonce`, `Actor()`, `SetActor()`.

## Handlers

* Handlers are registered per action type; the type key is `ActionObject.ObjectType()`.
* `TypedHandler` exposes `ActionType() string`.
* `Func[T ActionObject]` adapts a typed callback and type-asserts the incoming object to `T`, returning `false` on mismatch.

## Contract Delegation

If no local handler grants, `Authorize` consults active signed contracts:

* Look up active `SignedContract`s where `Subject == actor` and some permit `Action == action.ObjectType()`.
* For each match, swap `action.Actor` to `sc.Issuer` and re-run the local handlers.
* First `true` grants.

The contract path is purely a delegation: it does NOT call `Contract.Allows`, and `Constrainable.ApplyConstraints` is only used by callers that explicitly evaluate a contract against an action (e.g. `Contract.Allows`). Permit selection during authorization is the DB join in `findActiveContracts`.

## Contracts

`Contract`:

* Fields and co-signing semantics: see [mod.auth.contract](../../system/protocols/auth/types/mod.auth.contract.md).
* Implements `crypto.SignableTextObject`; signable hash is the contract's `ObjectID.Hash`.

`Permit`:

* Action name (`ObjectType` string) and optional `*astral.Bundle` constraints: see [mod.auth.permit](../../system/protocols/auth/types/mod.auth.permit.md).

`SignedContract`:

* Wraps `*Contract` with `IssuerSig` and `SubjectSig` (`*crypto.Signature`); both must be present (non-nil) and verify before indexing. Fields: see [mod.auth.signed_contract](../../system/protocols/auth/types/mod.auth.signed_contract.md).
* `IsNil()` is true when the embedded `*Contract` is nil.

## Constraints

`Constrainable` is an optional interface on actions: `ApplyConstraints(*Bundle) bool`. It is consulted by `Contract.Allows` when callers evaluate a contract against an action, but not by `Module.Authorize` itself.

## Built-Ins

`SudoAction`:

* Built-in action type.
* Requests permission for `Actor` to act as `AsID`.
* Registered authorizer grants only when `Actor.IsEqual(AsID)`; cross-identity sudo is reachable only through contract delegation.

`ContractQueryBuilder`:

* Fluent builder returned by `Module.SignedContracts()`.
* `WithIssuer(*Identity)`, `WithSubject(*Identity)`, `WithAction(...astral.Object)` (action filter is by `ObjectType()`).
* `Find(ctx)` returns active signed contracts (window: `starts_at <= now < expires_at`), decoding signatures and permits from `auth__contracts` + `auth__contract_permits`.

## Signing And Verification

* `Module.SignIssuer` / `Module.SignSubject` refuse to overwrite an existing signature with `ErrAlreadySigned`.
* `Module.SignContract` runs issuer then subject.
* `signAs` tries `Crypto.ObjectSigner` (`SchemeASN1`) first and falls back to `Crypto.TextObjectSigner` (`SchemeBIP137`).
* `VerifyContract` dispatches per `Signature.Scheme` accordingly.

## Object Holding

Active indexed signed-contract objects are held by `auth` against `objects.purge` via the `objects.Holder` hook, so authorization keeps working after a purge cycle. The hold window matches the active-contract lookup window.
