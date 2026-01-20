package objects

// Reader reads object opened by Read()
type Reader interface {
	Read(p []byte) (n int, err error)
	Close() error
	Repo() Repository
}
