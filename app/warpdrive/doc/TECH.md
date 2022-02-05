# Technical Details

Warpdrive consist two applications:

* Background service
    * Same implementation independent of operating system.
    * Communicates with UI client and other instances of warpdrive service.
    * Serves API for warpdrive UI client.
* UI client
    * Can differ depending on OS.
    * Allows the user:
        * sending files to other warpdrive users.
        * receiving notifications.
        * downloading files.

## Architecture

```
client[sender] <=> service[1] <=> service[2] <=> client[recipient]
```

## OS requirements

The operating system allows the application for:

* running ongoing background service written in golang.
* inter-process communication, web or unix sockets.
* displaying and updating the notification.
* access the file from user space as byte-stream.
* saving files in user space.

# API

The golang interfaces and structures for client API.

* [client.go](/app/warpdrive/api/client.go)
* [data.go](/app/warpdrive/api/data.go)

# Protocol

## Frames

| name       | short | format           | representation           | description                                                                                                                 |
|:-----------|:------|:-----------------|:-------------------------|-----------------------------------------------------------------------------------------------------------------------------|
| Recipients | rec   | struct           | []Peer                   | Detailed list of peers.                                                                                                     |
| Info       | info  | l8string, struct | OfferId, []Info          | Offer id and detailed files info. Contains information required by recipient to decide whatever accept or reject the offer. |
| Offer      | ofr   | struct           | Offer                    | Detailed collection of offers associated by ids.                                                                            |
| Offers     | ofs   | struct           | map[OfferId]Offer        | Detailed collection of offers associated by ids.                                                                            |
| Status     | stat  | struct           | Status                   | Offer status.                                                                                                               |
| Args       | arg   | l8string, struct | PeerId, []Info           | Offer arguments, contains peer id and path to file.                                                                         |
| Port       | port  | l8string         | string                   | Name of the port registered by sender, where the recipient can connect for downloading the requested files.                 |
| Id         | id    | l8string         | OfferId                  | Offer unique identifier.                                                                                                    |
| Attr       | attr  | 3x l8string      | [string, string, string] | Peer attribute for update.                                                                                                  |
| File       | file  | blob             | []byte                   | A file bytes.                                                                                                               |
| Close      | ok    | byte             | 0                        | Notifies connection is closing with success.                                                                                |

## Flow

| symbol | info                                                                        |    
|:------:|:----------------------------------------------------------------------------|
|   </   | open astral connection on port                                              |
|   <-   | send a frame                                                                |
|   ->   | receive a frame                                                             |
|   =>   | receive a stream of frames delayed in time                                  |
|   <>   | both sides can send or receive a frame, typically used for finishing stream |

### `sender`

Local protocol for communicating sender client with warpdrive service.

| </ `recipients` | </ `send` | </ `sent` | </ `events` |
|-----------------|-----------|-----------|-------------|
| -> rec          | <- arg    | -> ofs    | => stat     |
| <- ok           | -> ok     | <- ok     | <> ok       |

### `recipient`

Local protocol for communicating recipient client warpdrive service.

| </ `offers` | </ `accept` | </ `reject` | </ `received` | </ `events` | </ `update` |
|-------------|-------------|-------------|---------------|-------------|-------------|
| => ofr      | <- id       | <- id       | -> ofs        | => stat     | <- attr     |
| <> ok       | -> ok       | -> ok       | <- ok         | <> ok       | -> ok       |

### `service`

Remote protocol for communicating warpdrive services.

| </ `send` | </ `reject` | </ `accept` |
|-----------|-------------|-------------|
| <- info   | <- id       | <- id       |
| -> ok     | -> ok       | -> port     |
|           |             | <- query    |
|           |             | => file     |
|           |             | <- ok       |

# Persistent storage

* Sent offers
* Received offers
* Peers
* Files
