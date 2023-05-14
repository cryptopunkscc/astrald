package config

type PrefixStore struct {
	Store
	Prefix string
}

func NewPrefixStore(store Store, prefix string) *PrefixStore {
	return &PrefixStore{Store: store, Prefix: prefix}
}

func (s *PrefixStore) Read(name string) ([]byte, error) {
	return s.Store.Read(s.Prefix + name)
}

func (s *PrefixStore) Write(name string, data []byte) error {
	return s.Store.Write(s.Prefix+name, data)
}

func (s *PrefixStore) LoadYAML(name string, out interface{}) error {
	return s.Store.LoadYAML(s.Prefix+name, out)
}

func (s *PrefixStore) StoreYAML(name string, in interface{}) error {
	return s.Store.StoreYAML(s.Prefix+name, in)
}
