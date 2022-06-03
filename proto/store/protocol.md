# io:store protocol (draft)

io:store is a nestable, synchronous, asymmetric, binary protocol for managing sets of data blocks.

## Overview

| code | name   | desc                           |
|------|--------|--------------------------------|
| 0x01 | open   | open a block for reading       |
| 0x02 | create | create a new block for writing |

## Commands

### open (0x01)

Arguments

| type     | name  | desc                         |
|----------|-------|------------------------------|
| [40]byte | id    | id of the data block to open |
| uint32   | flags | flags (reserved)             |

Returned values

| type   | name     | desc                  |
|--------|----------|-----------------------|
| uint8  | error    | zero or an error code |

Error codes

| code | name        | desc                |
|------|-------------|---------------------|
| 0x01 | not found   | block not found     |
| 0xff | unavailable | command unavailable |

On success, a io:block protocol session begins.

### create (0x02)

Arguments

| type     | name  | desc                         |
|----------|-------|------------------------------|
| uint64   | alloc | allocate this many bytes     |

Returned values

| type   | name   | desc                  |
|--------|--------|-----------------------|
| uint8  | error  | zero or an error code |
| uint8  | id_len | temp id length        |
| []byte | id     | temp id               |

Error codes

| code | name        | desc                            |
|------|-------------|---------------------------------|
| 0x02 | no space    | not enough space for allocation |
| 0xfe | failed      | create failed                   |
| 0xff | unavailable | command unavailable             |

On success, a io:block protocol session begins.
