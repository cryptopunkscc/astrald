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
	SendingStatus(id RequestId) (string, error)
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
