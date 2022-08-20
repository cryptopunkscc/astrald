# publish event protocol (draft)
asynchronius, asymetric, protocol for publishing events

## Overview

|name|desc|visibility|
|-|-|-|
|publish|send event to all subscribers in channel|private|

## Commands

### publish
arguments

|type|name|desc|
|-|-|-|
|[c]c|channel|name of channel where evens are published|
|[c]c|data|data frame passed to channel|

returned values

|type|name|desc|
|-|-|-|
|c|error|zero or an error code|
