# shared content protocol (draft)
protocol for remote content access

## Overview

|name|desc|visibility|
|-|-|-|
|grant|grant the peer read access to the files|private|
|revoke|revoke the peer read access to the files|private|
|fetch|fetch identities of accessible files|public|
|sync|download the selected files from peer|public|
|close|close the connection|public|

## Commands

### grant
arguments

|type|name|desc|
|-|-|-|
|[]byte|peer|identity of peer to grant access|
|[][]byte|ids|list of files identities, that peer has access to|

returned values

|type|name|desc|
|-|-|-|
|byte|error|zero or an error code|

### revoke
arguments

|type|name|desc|
|-|-|-|
|[]byte|peer|identity of peer to revoke access|
|[][]byte|ids|list of files identities to revoke|

returned values

|type|name|desc|
|-|-|-|
|byte|error|zero or an error code|

### fetch
arguments

|type|name|desc|
|-|-|-|
|uint64|timestamp|timestamp of previous fetch|

returned values

|type|name|desc|
|-|-|-|
|[uint16][40]byte|ids|files identities|

### sync
arguments

|type|name|desc|
|-|-|-|
|[40]byte|id|file id|

returned values

|type|name|desc|
|-|-|-|
|byte|error|zero or an error code|

### close
returned values

|type|name|desc|
|-|-|-|
|0|ok|finalization byte|