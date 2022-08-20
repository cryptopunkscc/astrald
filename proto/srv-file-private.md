# private content protocol (draft)
protocol for local content management

## Note
consider to merging into [io:store](./store/protocol.md)

## Overview

|name|desc|visibility|
|-|-|-|
|list|get list of files identities|private|
|close|close the connection|public|

## Commands

### list
arguments

|type|name|desc|
|-|-|-|
|uint32|from|timestamp of the oldest element to include in results|

returned value

|type|name|desc|
|-|-|-|
|uint64|amount|amount of file identities|
|[amount]fileId|id|list of files identittes|

### close
returned values

|type|name|desc|
|-|-|-|
|0|ok|finalization byte|

## Dependencies

### [event:subscribe](mod-event-sub.md#subscribe)
required to collect identities into list

|channel|filter|data|type|
|-|-|-|-|
|block|0|id|[40]byte|