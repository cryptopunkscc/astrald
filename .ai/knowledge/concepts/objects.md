# Objects

## Core Types

### Object

* Immutable typed payload.
* Declares a type string.
* Serializes to and from bytes.
* Unknown types move forward as opaque blobs.

### ObjectID

* SHA-256 content address over type and payload.
* Wire format is 40 bytes: 8-byte size and 32-byte hash.
* String format is `data1` plus zBase32.
* Nodes compute ObjectIDs locally.

### Repository

* Stores raw bytes by ObjectID.
* Uses two-phase writes: get writer, then call `Commit()` or `Discard()`.

## Repository Groups

| Group       | Purpose                                        |
|-------------|------------------------------------------------|
| `local`     | primary on-disk; default write target          |
| `memory`    | in-memory cache                                |
| `removable` | portable/external media                        |
| `virtual`   | computed — archives, encryption wrappers       |
| `network`   | remote peers (requires Network zone)           |
| `system`    | internal node data                             |
| `main`      | device + virtual combined; default read target |

## Extension Points

Modules implement interfaces. The objects module discovers implementations
automatically through type assertions.

| Interface | Trigger |
|---|---|
| `Receiver` | remote push; accept and optionally persist |
| `Describer` | metadata request for an ObjectID |
| `Searcher` | text/tag search over module-owned indexes |
| `Finder` | provider lookup by ObjectID |
| `Holder` | storage policy and eviction decisions |
