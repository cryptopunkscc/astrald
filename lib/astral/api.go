package astral

import "io"

// Api interface for astral apps that have to be either standalone and embedded in node.
type Api interface {

	// Register new astral port under a given name.
	Register(name string) (Port, error)

	// Query a specific port by name. For calling a local service, pass empty string as nodeId.
	Query(nodeID string, query string) (io.ReadWriteCloser, error)

	// Resolve node identity by name.
	Resolve(name string) (string, error)
}

// Port for receiving local and remote requests.
type Port interface {

	// Next returns channel for receiving incoming requests
	Next() <-chan Request

	// Close and unregister the port.
	Close() error
}

// Request for new astral connection.
type Request interface {

	// Caller returns identity of callers node.
	Caller() string

	// Query returns the requested port name.
	Query() string

	// Accept incoming connection and start the stream.
	Accept() (io.ReadWriteCloser, error)

	// Reject the incoming request.
	Reject()
}
