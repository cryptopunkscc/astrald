package resources

type Resources interface {
	Read(name string) ([]byte, error)
	Write(name string, data []byte) error
}
