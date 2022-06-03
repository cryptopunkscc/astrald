# io:block protocol (draft)

io:block is a nestable, synchronous, asymmetric, binary protocol that gives random access to a data block.

## Overview

The client sends a command followed by its arguments. Commands are single byte values: 

| code | name     | desc                                   |
|------|----------|----------------------------------------|
| 0x01 | read     | read bytes                             |
| 0x02 | write    | write bytes                            |
| 0x03 | seek     | seek to a position                     |
| 0x04 | finalize | finalize data and get its immutable id |
| 0xff | end      | ends the protocol                      |

The client should wait for server response before sending the next command.

## Commands

### read (0x01)

Arguments

| type   | name      | desc                               |
|--------|-----------|------------------------------------|
| uint16 | max_bytes | maximum number of bytes to be read |

Returned values

| type   | name     | desc                  |
|--------|----------|-----------------------|
| uint8  | error    | zero or an error code |
| uint16 | data_len | data length           |
| []byte | data     | data                  |

Error codes

| code | name        | desc                |
|------|-------------|---------------------|
| 0x01 | eob         | end of block        |
| 0xfe | failed      | read failed         |
| 0xff | unavailable | command unavailable |

#### Notes

If the argument max_bytes is 0, the request is still valid, the host should return no error and null data.

An error response can still contain data.

### write (0x02)

Arguments

| type   | name     | desc        |
|--------|----------|-------------|
| uint16 | data_len | data length |
| []byte | data     | data        |

Returned values

| type   | name  | desc                    |
|--------|-------|-------------------------|
| uint8  | error | zero or an error code   |
| uint16 | count | number of bytes written |

Error codes

| code | name        | desc                |
|------|-------------|---------------------|
| 0x02 | nospace     | no space left       |
| 0xfe | failed      | write failed        |
| 0xff | unavailable | command unavailable |

### seek (0x03)

Arguments

| type  | name   | desc                            |
|-------|--------|---------------------------------|
| int64 | offset | target offset relative to ref   |
| uint8 | ref    | 0 - start, 1 - current, 2 - end |

Returned values

| type   | name  | desc                      |
|--------|-------|---------------------------|
| uint8  | error | zero or an error code     |
| uint64 | pos   | new position in the block |

Error codes

| code | name        | desc                |
|------|-------------|---------------------|
| 0xfe | failed      | seek error          |
| 0xff | unavailable | command unavailable |

### finalize (0x04)

This command has no arguments.

Returned values

| type     | name  | desc                           |
|----------|-------|--------------------------------|
| uint8    | error | zero or an error code          |
| [40]byte | id    | id of the finalized data block |

Error codes

| code | name        | desc                |
|------|-------------|---------------------|
| 0xff | unavailable | command unavailable |

### end (0xff)

End takes no arguments and rerturns a single null byte which is the final byte of the protocol.
