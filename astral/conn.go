package astral

// Conn defines the basic interface of an astral connection.
// Beyond io semantics, both endpoints are identity-bound: every Conn exposes the
// authenticated local and remote identities of the link.
type Conn interface {
	Read(b []byte) (n int, err error)
	Write(b []byte) (n int, err error)
	Close() error
	LocalIdentity() *Identity
	RemoteIdentity() *Identity
}
