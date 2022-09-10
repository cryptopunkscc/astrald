# warp drive cli

Command line interface for warpdrive.

## Run

To start interactive mode, run following command from repository root: 

```shell
go run ./cmd/warpdrive
```

To run single command and exit, append command name and required arguments, for example:  

```shell
go run ./cmd/warpdrive get 9c10b4cb-b770-4602-4b51-0e755b1cea2a
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
