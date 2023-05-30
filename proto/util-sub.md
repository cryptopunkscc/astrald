# subscribe protocol (draft)
asynchronius, asymetric, protocol for subscribing events

## Overview

|name|desc|visibility|
|-|-|-|
|subscribe|subscribe to event channel|private|

## Commands

### subscribe
arguments

|type|name|desc|
|-|-|-|
|[c]c|name|name of subscriber|
|\*|filter|meta-data values filter|
|c|pulse|pulse signal for keeping connection alive|

returned values

|type|name|desc|
|-|-|-|
|c|code|zero or an error code|
|\*|data|data frame passed to channel|

flow:
1. send `channel` name
2. receive `code`
	1. on error return
3. spawn async workers
	1. inside loop, send `pulse` signal and wait a delay
	2. inside loop, read `data` frames
