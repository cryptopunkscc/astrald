package fs

type Config struct {
	Watch  []string      // list of paths to index for read-only storage
	Store  []string      // list of paths to use for read-write storage
	Shares []shareConfig // list of dirs to share with other identities
}

type shareConfig struct {
	Path  string
	Allow []string
}

var defaultConfig = Config{}
