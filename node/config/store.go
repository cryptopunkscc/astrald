package config

type Store interface {
	Read(name string) ([]byte, error)
	Write(name string, data []byte) error
	LoadYAML(name string, out interface{}) error
	StoreYAML(name string, in interface{}) error
	BaseDir() string
}
