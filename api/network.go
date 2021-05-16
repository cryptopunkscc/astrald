package api

// Network provides access to core network APIs
type Network interface {
	Register(name string) (PortHandler, error)
	Connect(identity Identity, port string) (Stream, error)
	Identity() Identity
}

// PortHandler is a handler for a locally registered port
type PortHandler interface {
	Requests() <-chan ConnectionRequest
	Close() error
}

// ConnectionRequest represents a connection request sent to a port
type ConnectionRequest interface {
	Caller() Identity
	Query() string
	Accept() Stream
	Reject()
}

// Stream represents a bidirectional stream
type Stream interface {
	Read(p []byte) (n int, err error)
	Write(p []byte) (n int, err error)
	Close() error
}

type Identity string
