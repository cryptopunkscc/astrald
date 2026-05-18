# Wire

## Encodings

| Format    | Framing                                     | Used for                  |
|-----------|---------------------------------------------|---------------------------|
| binary    | `String8(type)` + `Bytes32(payload)`        | default channel transport |
| json      | `{"Type":"...","Object":{...}}\n`           | debugging, `astral-query` |
| text      | `#[type] value\n` or `#[type]:base64\n`     | human-readable output     |
| canonical | `Stamp(4b)` + `String8(type)` + raw payload | storage, ObjectID hashing |

Invariant: only canonical encoding produces a stable `ObjectID`. Do not use binary framing for storage.

## Objectify Fields

`astral.Objectify` reflects struct fields in declaration order.

Supported field types:

* `String8/16/32`
* numeric types: `Int8`...`Int64`, `Uint8`...`Uint64`
* `Bool`, `Float32/64`, `Duration`, `Nonce`, `*Identity`, `ObjectID`
* pointers, slices, arrays, and maps of supported types

Trust boundary: interface fields of type `astral.Object` use dynamic framing: type name plus payload.
