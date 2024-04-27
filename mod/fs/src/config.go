package fs

type Config struct {
	Watch []string // list of paths to index for read-only storage
	Store []string // list of paths to use for read-write storage
}

var defaultConfig = Config{}
