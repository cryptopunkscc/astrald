# mod/auth

Decides whether an identity may perform a typed action, using local action handlers first and signed delegation contracts as a fallback. Owns the authorization handler registry, signed-contract verification and indexing, and the persistent contract permit index used by authorization lookups.

## Dependencies

| Module | Why |
| --- | --- |
| `crypto` | signs contracts through `ObjectSigner` and `TextObjectSigner`; verifies issuer and subject signatures with ASN1 or BIP137 schemes |
| `objects` | loads signed contracts for `auth.index`; scans `RepoLocal` for startup indexing; exposes active-contract object holds for purge |
| `secp256k1` | derives issuer and subject public keys from identities for signing and verification |
| `core/assets` | `Database()` backs `auth__contracts` and `auth__contract_permits`; `LoadYAML` loads the empty auth config |

## Flows

- Direct authorization: `Authorize` reads `action.ObjectType()` and `action.Actor()` -> tries registered handlers for that action type -> first handler returning true grants the action.
- Contract authorization: handler miss -> `SignedContracts().WithSubject(actor).WithAction(action).Find` -> for each active contract, replace the action actor with `sc.Issuer` -> retry local handlers.
- Register handler: `Module.Add` reads each `TypedHandler.ActionType()` -> appends it in the in-memory handler map keyed by object type.
- Sudo handler: loader registers `auth.Func[*auth.SudoAction](AuthorizeSudo)` -> `AuthorizeSudo` grants only when `Actor` equals `AsID`.
- Sign contract op: `auth.sign_contract` accepts a `Contract` from the query channel -> wraps it in `SignedContract` -> `SignIssuer` -> `SignSubject` -> sends the signed object or an error object.
- Sign one side: `SignIssuer` or `SignSubject` refuses an existing signature -> derives the identity public key with `secp256k1.FromIdentity` -> tries ASN1 object signing -> falls back to BIP137 text-object signing.
- Index contract op: `auth.index` loads the object ID through `Objects.ReadDefault()` -> requires `*SignedContract` -> `IndexContract` verifies and stores it -> sends `Ack`.
- Index contract: `indexMu` -> resolve object ID -> skip when a complete contract row already exists -> verify issuer and subject signatures according to signature scheme -> upsert contract row and insert permits if no permits exist yet.
- Object holding: `HoldObject` returns true for active indexed signed-contract object IDs so `objects.purge` skips contracts still used for authorization.
- Startup indexer: `Run` starts `indexer` -> scan `objects.RepoLocal` outside `ZoneNetwork` -> load each object -> index signed contracts and ignore other objects.
- Query contracts: `SignedContracts` builder filters by issuer, subject, active time window, and permit action type -> decodes stored signatures and permits back into `SignedContract` objects.

## Source

- `mod/auth/module.go`, `action.go`, `actions_map.go`, `contract.go`, `signed_contract.go`, `sudo_action.go`, `errors.go` - public authorization interfaces, action and contract types, sudo action, and sentinels.
- `mod/auth/src/loader.go`, `module.go`, `deps.go`, `config.go`, `prepare.go` - module registration, dependency injection, router setup, database migration, and lifecycle.
- `mod/auth/src/authorize.go`, `authorizers.go` - handler dispatch, contract fallback, and built-in sudo authorization.
- `mod/auth/src/signing.go` - issuer and subject signing plus signature verification by scheme.
- `mod/auth/src/contracts.go`, `contract_query.go`, `object_holder.go` - contract indexing, repository scan, active-contract query builder, and cleanup hold hook.
- `mod/auth/src/op_sign_contract.go`, `op_index.go` - query operation handlers.
- `mod/auth/src/db.go`, `db_contract.go`, `db_contract_permit.go` - GORM rows, contract upsert, permit persistence, and lookup filters.
- `mod/auth/client/` - typed client wrappers for auth operations.

## Surface

| What | Why it matters |
| --- | --- |
| `Module.Authorize` and `Module.Add` | central extension point for modules that define typed authorization actions |
| `Module.SignContract`, `VerifyContract`, and `IndexContract` | contract lifecycle used by modules that delegate authority between identities |
| `Module.SignedContracts()` | builder used by authorization and callers that need active signed contract lookup |
| `objects.Holder` | prevents active indexed signed-contract objects from being purged |
| `auth.sign_contract`, `auth.index` | query methods for signing a contract object and indexing an already stored signed contract |
| `auth__contracts`, `auth__contract_permits` | persistent authorization index searched during contract fallback |

## Invariants

- Contracts never grant directly; they swap `action.Actor` to issuer and re-run local handlers.
- Contract path skips `Contract.Allows` and `Constrainable.ApplyConstraints`.
- Contract lookup filters `subject_id`, `starts_at <= now`, `expires_at` zero-or-future, joins permits by action `ObjectType`.
- `HoldObject` uses the same active time window as contract lookup and fails closed on DB errors.
- `contractExists` requires both signatures non-empty; partial rows re-index.
- `IndexContract` serialized by `indexMu`.
- `SignIssuer`/`SignSubject` refuse overwrite with `ErrAlreadySigned`.
- `auth.yaml` has no fields; presence is inert.
