package fs

const DefaultRepoName = "data"

type Config struct {
	Repos map[string]RepoConfig
}

type RepoConfig struct {
	Label    string
	Path     string
	Writable bool
}

var defaultConfig = Config{}
