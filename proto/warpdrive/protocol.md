# warpdrive protocol (draft)

warpdrive is a binary protocol for sharing files over astral network.

## Overview

Common

| code | name          | desc                                                |
|------|---------------|-----------------------------------------------------|
| 0xff | end           | end session                                         |

Local node

| code | name          | desc                                                  |
|------|---------------|-------------------------------------------------------|
| 0x01 | list peers    | list known peers                                      |
| 0x02 | create offer  | create offer with info about files available for peer |
| 0x03 | accept offer  | accept received and start downloading in background   |
| 0x04 | list offers   | list all incoming and/or outgoing offers              |
| 0x04 | listen offers | listen for incoming and/or outgoing offers            |
| 0x05 | listen status | listen status of incoming and/or outgoing offers      |

Remote node

| code | name       | desc                                    |
|------|------------|-----------------------------------------|
| 0x01 | send offer | send info about available files to peer |
| 0x02 | download   | download file from specific offer       |

Info

| code | name | desc                       |
|------|------|----------------------------|
| 0x01 | ping | ping the warpdrive service |

## Local commands

### list peers

Returned values

| type   | name     | desc                  |
|--------|----------|-----------------------|
| []Peer | error    | zero or an error code |

### create offer

Arguments

| type | name      | desc                              |
|------|-----------|-----------------------------------|
| [c]c | peer id   | id of peer node                   |
| [c]c | file path | path to file or directory to send |

Returned values

| type | name        | desc                       |
|------|-------------|----------------------------|
| [c]c | offer id    | id of files offers         |
| c    | result code | 0 - awaiting, 1 - accepted |

### accept offer

Arguments

| type | name     | desc               |
|------|----------|--------------------|
| [c]c | offer id | id of files offers |

Returned values

| type | name | desc                  |
|------|------|-----------------------|
| c    | code | zero or an error code |

### list offers

Arguments

| type | name   | desc                                |
|------|--------|-------------------------------------|
| c    | filter | 0 - all, 1 - incoming, 2 - outgoing |

Returned values

| type    | name         | desc                  |
|---------|--------------|-----------------------|
| []Offer | offers list  | zero or an error code |

### listen offers

Arguments

| type | name   | desc                                |
|------|--------|-------------------------------------|
| c    | filter | 0 - all, 1 - incoming, 2 - outgoing |

Stream values

| type  | name      | desc                  |
|-------|-----------|-----------------------|
| Offer | new offer | zero or an error code |

Finalize

| type | name | desc                |
|------|------|---------------------|
| 0    | code | finish subscription |

### listen status

Arguments

| type | name   | desc                                |
|------|--------|-------------------------------------|
| c    | filter | 0 - all, 1 - incoming, 2 - outgoing |

Returned stream

| type        | name         | desc                  |
|-------------|--------------|-----------------------|
| OfferStatus | offer status | zero or an error code |

Finalize

| type | name | desc                |
|------|------|---------------------|
| 0    | code | finish subscription |

## Remote commands

### send offer

Arguments

| type   | name       | desc                                  |
|--------|------------|---------------------------------------|
| [c]c   | offer id   | the offer id                          |
| []Info | files info | files info associated to the offer id |

Returned value

| type | name        | desc                       |
|------|-------------|----------------------------|
| c    | result code | 0 - awaiting, 1 - accepted |

### download

Arguments

| type | name     | desc                      |
|------|----------|---------------------------|
| [c]c | offer id | the offer id              |
| q    | index    | index of file to download |
| q    | offset   | offset of file            |

Returned values

| type | name  | desc              |
|------|-------|-------------------|
| c    | code  | 0 - confirmation  |
| blob | files | files byte stream |

Finalize

| type | name | desc               |
|------|------|--------------------|
| 0    | code | finish downloading |

### ping

Arguments

| type | name   | desc                      |
|------|--------|---------------------------|
| c    | signal | 0 - close, (c > 0) - ping |

Returned values

| type | name   | desc                     |
|------|--------|--------------------------|
| c    | signal | byte value from argument |
