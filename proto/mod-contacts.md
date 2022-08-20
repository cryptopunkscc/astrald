# contacts protocol (draft)

contacts is a asynchronous protocol for retriving contacts details

## Overview

|name|desc|visibility|
|-|-|-|
|list|obtain contacts list|private|

## Data

### contact

|type|name|desc|
|-|-|-|
|[c]c|name|key used to obtain the metadata parameter value|
|[33]c|id|regex pattern for matching meta-data value|


## Commands

### list
returned values

|type|name|desc|
|-|-|-|
|[]contact|contacts|list of known contacts|

## Side effects

### new contact -> [event:publish](mod-event-pub.md#publish)
any newly recognised contact is published to event channel 

|channel|data|type|
|-|-|-|
|contact|contact|[uint16]byte|