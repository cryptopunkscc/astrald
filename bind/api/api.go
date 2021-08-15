package astralApi

type Service interface {
	Run(core Network) error
}

// Network provides access to core network APIs
type Network interface {
	Register(name string) (PortHandler, error)
	Connect(identity string, port string) (Stream, error)
	Identity() string
}

// PortHandler is a handler for a locally registered port
type PortHandler interface {
	Next() ConnectionRequest
	Close() error
}

// ConnectionRequest represents a connection request sent to a port
type ConnectionRequest interface {
	Caller() string
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
