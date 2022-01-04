# Warp Drive v1.0.0-draft

Application for sending files over the astral network. Comes as two separated parts:

* Background service
    * Same implementation independent of OS.
    * Connects to node through apphost module.
    * Communicates with other instances of warp drive service.
    * Serves API for warp drive UI client
    * Can be embedded into node or provided as standalone application.
* UI Application
    * Can differ depending on OS.
    * Standalone application.
    * Allows sending files to other warp drive users over the astral network receiving notifications, accepting and
      rejecting incoming files.
    * The user can act with UI as a sender or recipient.

## User stories

* As a `sender`, I can:
    1. see the list of `recipients`.
    2. `send` file to the recipient.
    3. see the list of `sent` requests.
    4. be notified about outgoing files status `events`.
* As a `recipient`, I can:
    1. be notified about `incoming` files requests.
    2. `accept` incoming file request.
    3. `reject` incoming file request.
    4. `update` peer as: 
       1. `trusted` to automatically accept incoming files requests.
       2. `blocked` to automatically reject incoming files requests.
    5. see the list of `received` requests.
    6. be notified about incoming files status `events`.

## Client API

```go
package warpdrive

import "os"

type UIApi interface {
	SenderApi
	RecipientApi
	Sender() SenderApi
	Recipient() RecipientApi
}

type SenderApi interface {
	StatusApi
	// Peers lists available recipients.
	Peers() ([]Peer, error)
	// SendFile sends files request to the recipient.
	SendFile(peerId string, path string) (RequestId, error)
	// SentRequests returns collection of sent files requests.
	SentRequests() (map[RequestId]OutgoingFiles, error)
}

type RecipientApi interface {
	StatusApi
	// IncomingFiles subscribes a callback for receiving incoming files requests.
	IncomingFiles() (<-chan IncomingFiles, error)
	// ReceivedRequests returns collection of received files requests
	ReceivedRequests(filterStatus string) (map[RequestId]IncomingFiles, error)
	// AcceptRequest accepts incoming files and starts downloading.
	AcceptRequest(id RequestId) error
	// RejectRequest rejects incoming files requests.
	RejectRequest(id RequestId) error
	// UpdatePeer updates peer attribute [alias, mod].
	UpdatePeer(id PeerId, attr string, value string) error
}

type StatusApi interface {
	// Events subscribes a callback for receiving request status updates.
	Events() (<-chan RequestStatus, error)
}

type Id string

type RequestId Id

type OutgoingFiles struct {
	FilesRequest
	Recipient PeerId
}

type IncomingFiles struct {
	FilesRequest
	Sender PeerId
}

type FilesRequest struct {
	RequestStatus
	Files []FileInfo
}

type RequestStatus struct {
	Id     RequestId
	Status string
}

type FileInfo struct {
	Path  string
	Size  int64
	IsDir bool
	Perm  os.FileMode
	Mime  string
}

type PeerId string

type Peer struct {
	Id       PeerId
	Hostname string
	Alias    string
	Mod      string
}

const (
	PEER_MOD_ASK   = ""
	PEER_MOD_TRUST = "trust"
	PEER_MOD_BLOCK = "block"
)

```

## Persistent storage

* Outgoing files requests
* Incoming files requests
* Peers
* Files

## Sender protocol

Local protocol for communicating sender client with warpdrive service.

### `recipients`

1. <- Query recipients
2. -> Recipients
3. <- OK

### `send`

1. <- Query sender send
2. <- Send recipient id and file path
    * Invalid id!
    * path!
3. -> OK

### `sent`

1. <- Query sender sent
2. -> Requests
3. <- OK

### `events`

1. <- Query sender events
2. ->> Event
3. <-> OK

## Recipient protocol

Local protocol for communicating recipient client warpdrive service.

### `incoming`

1. <- Query recipient incoming
2. ->> Incoming files
3. <-> OK

### `accept`

1. <- Query recipient accept
2. <- Request id
3. -> OK

### `reject`

1. <- Query recipient reject
2. <- Request id
3. -> OK

### `update`

1. <- Query recipient update
2. <- Recipient mod
3. -> OK

### `received`

1. <- Query recipient received
2. -> Received files requests
3. <- OK

### `events`

1. <- Query sender events
2. ->> Event
3. <-> OK

## Service protocol

Remote protocol for communicating warpdrive services.

### `send`

1. <- Query send
    * Peer blocked!
2. <- Files request body
3. -> OK

### `accept`

1. <- Query accept
2. <- Accepted files request id
    * Invalid id!
3. -> Files port name
4. <- Query files
5. -> Files
6. <- OK

### `reject`

1. <- Query reject
2. <- Rejected files request id
    * Invalid id!
3. -> OK
