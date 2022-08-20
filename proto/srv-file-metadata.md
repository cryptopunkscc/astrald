# content metadata protocol (draft)
protocol for accesing files metadata

## Overview

|name|desc|visibility|
|-|-|-|
|query|obtain filtered metadata cursor|private|

## Commands

### query
returned values

|type|name|desc|
|-|-|-|
|byte|error|zero or an error code|

On success, a [util-cursor](util-cursor.md) protocol session begins, providing metadata items.

### close
returned values

|type|name|desc|
|-|-|-|
|0|ok|finalization byte|

## Side effects

### new metadata -> [event:publish](mod-event-pub.md#publish)
any newly recognised metadata is published to event channel 

|channel|data|type|
|-|-|-|
|meta|metadata|[uint16]byte|

### new blob -> [event:publish](mod-event-pub.md#publish)
any newly recognised blob id is published to event channel 

|channel|data|type|
|-|-|-|
|blob|id|[40]byte|
