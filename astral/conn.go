package astral

// Conn defines the basic interface of an astral connection
type Conn interface {
	Read(b []byte) (n int, err error)
	Write(b []byte) (n int, err error)
	Close() error
	LocalIdentity() *Identity
	RemoteIdentity() *Identity
}
