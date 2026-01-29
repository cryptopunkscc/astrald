package crypto

const maxObjectSize = 4096

type Config struct {
	Repos []string
}

var defaultConfig = Config{
	Repos: []string{"local", "mem0", "system"},
}
