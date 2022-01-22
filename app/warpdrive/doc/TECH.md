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

```go
package warpdrive

import "os"

type ClientApi interface {
	SenderApi
	RecipientApi
	Sender() SenderApi
	Recipient() RecipientApi
}

type SenderApi interface {
	StatusApi
	// Peers available for receiving an offer.
	Peers() ([]Peer, error)
	// Send files offer for the recipient.
	Send(peerId PeerId, files []Info) (OfferId, error)
	// Sent offers.
	Sent() (Offers, error)
	Status(id OfferId) (string, error)
}

type RecipientApi interface {
	StatusApi
	// Offers subscription.
	Offers() (<-chan Offer, error)
	// Received offers.
	Received() (Offers, error)
	// Accept offer and starts in background downloading.
	Accept(id OfferId) error
	// Reject offer.
	Reject(id OfferId) error
	// Update peer attribute [alias, mod].
	Update(id PeerId, attr string, value string) error
}

type StatusApi interface {
	// Events subscribes a callback for receiving request status updates.
	Events() (<-chan Status, error)
}

// =================== Peer ===================

type Peers map[PeerId]*Peer
type PeerId string

type Peer struct {
	Id    PeerId
	Alias string
	Mod   string
}

// =================== Offer ===================

type Offers map[OfferId]*Offer
type OfferId string

type Offer struct {
	Status
	Peer  PeerId
	Files []Info
}

type Status struct {
	Id     OfferId
	Status string
}

type Info struct {
	uri  string
	Size  int64
	IsDir bool
	Perm  os.FileMode
	Mime  string
}

const (
	STATUS_SENT = iota
	STATUS_RECEIVED
	STATUS_ACCEPTED
)

const (
	PEER_MOD_ASK   = ""
	PEER_MOD_TRUST = "trust"
	PEER_MOD_BLOCK = "block"
)
```

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
