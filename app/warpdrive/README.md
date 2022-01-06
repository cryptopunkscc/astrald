# Warp Drive v1.0.0-draft

Application for sending files over the astral network. Comes as two separated parts:

* Background service
    * Same implementation independent of OS.
    * Connects to node through `apphost` module.
    * Communicates with other instances of warp drive service.
    * Serves API for warp drive UI client
    * Can be embedded into node or provided as standalone application.
* UI Application
    * Can differ depending on OS.
    * Standalone application.
    * Allows sending files to other warp drive users over the astral network as well as receiving notifications,
      accepting and rejecting incoming files.
    * The user can act with UI as a sender or recipient.

## User stories

* As a `sender`, I can:
    1. see the list of `recipients`.
    2. `offer` files for the recipient.
    3. see the list of offers that I `sent` to the recipient.
    4. be notified about status `events` about offers that I sent.
* As a `recipient`, I can:
    1. be notified about incoming `offers`.
    2. `accept` offer.
    3. `reject` offer.
    4. `update` peer as:
        1. `trusted` to automatically accept incoming offers.
        2. `blocked` to automatically reject incoming offers.
    5. see the list of `received` requests.
    6. be notified about incoming files status `events`.

## API

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
	// Recipients available for receiving an offer.
	Recipients() ([]Peer, error)
	// Send files offer for the recipient.
	Send(peerId PeerId, path string) (OfferId, error)
	// Sent offers.
	Sent() (map[OfferId]Offer, error)
}

type RecipientApi interface {
	StatusApi
	// Offers subscription for receiving incoming requests.
	Offers() (<-chan Offer, error)
	// Received offers.
	Received(filterStatus string) (map[OfferId]Offer, error)
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

type Offer struct {
	Status
	Peer  PeerId
	Files []Info
}

type Status struct {
	Id     OfferId
	Status string
}

type OfferId string

type Peer struct {
	Id    PeerId
	Alias string
	Mod   string
}

type PeerId string

type Info struct {
	Path  string
	Size  int64
	IsDir bool
	Perm  os.FileMode
	Mime  string
}

const (
	PEER_MOD_ASK   = ""
	PEER_MOD_TRUST = "trust"
	PEER_MOD_BLOCK = "block"
)

```

## Architecture

```
client[sender] <=> service[1] <=> service[2] <=> client[recipient]
```

## Frames

| name       | short | format      | representation           | description                                                                                                    |
|:-----------|:------|:------------|:-------------------------|----------------------------------------------------------------------------------------------------------------|
| Recipients | rec   | struct      | []Peer                   | Detailed list of peers.                                                                                        |
| Info       | info  | struct      | []Info                   | Detailed files info. Contains information required by recipient to decide whatever accept or reject the offer. |
| Offer      | ofr   | struct      | Offer                    | Detailed collection of offers associated by ids.                                                               |
| Offers     | ofs   | struct      | map[OfferId]Offer        | Detailed collection of offers associated by ids.                                                               |
| Status     | stat  | struct      | Status                   | Offer status.                                                                                                  |
| Args       | arg   | 2x l8string | [string, string]         | Offer arguments, contains peer id and path to file.                                                            |
| Port       | port  | l8string    | string                   | Name of the port registered by sender, where the recipient can connect for downloading the requested files.    |
| Id         | id    | l8string    | OfferId                  | Offer unique identifier.                                                                                       |
| Attr       | attr  | 3x l8string | [string, string, string] | Peer attribute for update.                                                                                     |
| File       | file  | blob        | []byte                   | A file bytes.                                                                                                  |
| Close      | ok    | byte        | 0                        | Notifies connection is closing with success.                                                                   |

## Protocol

|  query  |  send  |  receive  |  stream  |  send/receive  |
|:-------:|:------:|:---------:|:--------:|:--------------:|
|   </    |   <-   |    ->     |    =>    |       <>       |

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

## Persistent storage

* Sent offers
* Received offers
* Peers
* Files
