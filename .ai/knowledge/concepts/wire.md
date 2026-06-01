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

`astral.Objectify(&v)` reflects a value into binary, JSON, and a derived
`ObjectType()`. Struct fields are read and written in declaration order.

Supported kinds:

* numeric: `Int8`...`Int64`, `Uint8`...`Uint64`, `Float32/64`
* `Bool`, `String`, `Duration`, `Nonce`, `*Identity`, `ObjectID`, `Zone`
* `Ptr`, `Slice`, `Array`, `Map`, nested `Struct`, `Interface`

Interface fields typed as `astral.Object` use dynamic framing (type name +
payload). Types decoded into them must be registered with `astral.Add`, kept
in a `Blueprints` registry.

## Helpers

* `astral.Stringify(v)` returns `Stringer.String()`, falls back to
  `TextMarshaler.MarshalText`, then `%v`.
* `astral.New(typeName)` returns a zero-value object from the default
  `Blueprints` or nil.
