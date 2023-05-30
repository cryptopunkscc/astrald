# drive (draft) 
application protocol for file management

## Overview

|name|desc|
|-|-|
|pull|pull resources from given uri into storage|
|fetch|similar to pull but saves only metadata and links to encountered blobs|
|read|open read stream for uri|
|share|make specified resources available to read for a given peer|
|revoke|oposite to share|
|mount|makes view to specified resources in native file system|

## Commands

### pull
arguments

|type|name|desc|
|-|-|-|
|[c]c|uri|universal resource identifier to file|

returned values

|type|name|desc|
|-|-|-|
|[c][40]c|ids|files identities|

### fetch
arguments

|type|name|desc|
|-|-|-|
||||

returned values

|type|name|desc|
|-|-|-|
||||

### read
arguments

|type|name|desc|
|-|-|-|
||||

returned values

|type|name|desc|
|-|-|-|
||||

### read
arguments

|type|name|desc|
|-|-|-|
||||

returned values

|type|name|desc|
|-|-|-|
||||