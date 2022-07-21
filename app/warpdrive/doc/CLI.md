# Warp Drive CLI v1.0.0-dev

Command line service for warpdrive, combined with [ANC](../../../cmd/anc/README.md) (astral netcat) can be considered as
referential UI client for development and testing purpose. Under the hood, it translates text commands directly into
warpdrive UI API calls and returns received results as formatted console output. Depending on needs is recommended to
use more than one client instance connected at the same time, this can be useful for tracking subscription updates and
triggering updates at the same time.

### Running console client

ANC is a commandline astral client, that allows querying services registered under specified port, on local or remote
device:

```shell
$ anc query <identity> <port>
```

The warpdrive CLI allows only for localhost connections, so you have to omit the identity as redundant, and leave only
the correct warpdrive port.

For example, to connect to warpdrive CLI, you can run ANC directly from source file [main.go](../../../cmd/anc/main.go):

```shell
$ go run ./cmd/anc/main.go query wd
connected.
```

## Commands

* peers - Print list of peers  
* send <file_path> <peer_id>? - Send a file or directory.  
* out - Print list of outgoing files.
* in - Print list of incoming files.
* sub <all|in|out>? - Subscribe for receiving new file offers.
* get <offer_id> - Download the files from specific offer in background. 
* stat <all|in|out>? - Subscribe for offers status updates. 
* update - Update specific attribute of peer. 
  * alias <peer_alias> - Change peer alias
  * mod <trust|block>? - Auto download or reject files from peer. Leave empty string for default.
