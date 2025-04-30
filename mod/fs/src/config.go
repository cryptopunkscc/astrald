package fs

type Config struct {
	Repos map[string]string // list of paths to use for read-write storage
	Watch map[string]string // list of paths to use for read-only storage
}

var defaultConfig = Config{}
