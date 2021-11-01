package storage

type Store interface {
	StoreBytes(file string, data []byte) error
	LoadBytes(file string) ([]byte, error)
}
