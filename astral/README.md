# astral

This package documents and implements core astral components.

## Definitions

### object

An object is a [payload](#payload) with an optional [type](#type).
An object without a type is called an [anonymous object](#anonymous-object).
All objects are immutable and have a [canonical form](#canonical-form).

### payload

Any binary data.

### type

A type is a non-empty string of up to 255 characters. A type is used to convey
how to interpret the data in the payload. Allowed characters are alphanumeric,
period, hyphen and underscore.

### implicit typing

If the object type can be unambiguously inferred from the context, its payload
is enough to correctly interpret its content, and its type and does not need
to be explicitly stated.

### object id

An object id is an uint64 representing the size of the object in bytes
followed by SHA256 hash of the [canonical form](#canonical-form) of the object.
In effect, object id is a 320-bit value uniquely identifying an object. Object
id is itself an object of type `astral.object_id.sha256`.

### anonymous object

An anonymous object is an object that has no type. Any binary data can be
treated as an anonymous object. Anonymous objects can still be correctly
interpreted if they can be implicitly typed and their structure can be
inferred from the context.

### canonical form

A canonical form of an object is its header followed by its payload.

### header

A binary representation of a [type](#type). If an object has no [type](#type),
the header is empty and has zero length. Otherwise, the header consists of the
magic bytes, followed by an uint8 specifying type length, followed by the type
encoded as an ASCII string. The header itself is an object with a type of
`astral.object_header`.

### short object header

An object header without the magic bytes.

### magic bytes

A const uint32 value equal to `0x41444330`.

### identity

Identity is a secp256k1 public key serialized into a 33-byte long compressed
format. Identity is an object of the type `astral.identity.secp256k1`.

### zone

A zone describes what resources need to be used to access an object or an
identity. There are 3 zones: device, virtual and network (shortened to d, v
and n respectively).

#### device

All objects stored on the local device and the pheripherals directly attached
to it. All local identities such as the node and its guests.

#### virtual

Objects that are not stored in their original representation, but can be
generated on demand (ex. extracted from a zip file).

#### network

Objects and identities accessbile via a network.