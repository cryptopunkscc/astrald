# cursor protocol (draft)
protocol for iterating over large data set

## Overview

|name|desc|visibility|
|-|-|-|
|filter|setup byte stream filter|private|
|next|get specified amount of items and move offset|private|
|seek|skip specified amount of items and move offset|private|
|close|close the cursor|private|

## Commands

### filter
arguments

|type|name|desc|
|-|-|-|
|[filter](./util-filter#filter)|filter|meta-data values filter|

returned values

|type|name|desc|
|-|-|-|
|c|error|zero or an error code|

### next
arguments

|type|name|desc|
|-|-|-|
|int32|amount|amont of items to get|

returned values

|type|name|desc|
|-|-|-|
|int32|amount|amount of items to return, can be lower then requested amount|
|[][]byte|items|array of encoded items|

#### Note
Amount can be negative for navigating backward.

### seek
arguments

|type|name|desc|
|-|-|-|
|int32|amount|amont of items to skip|

returned values

|type|name|desc|
|-|-|-|
|int32|amount|amount of items to return|

#### Note
Amount can be negative for navigating backward.

### close
returned values

|type|name|desc|
|-|-|-|
|0|ok|finalization byte|
