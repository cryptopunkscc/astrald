package zip

type Config struct {
	Virtual bool // automatically index zip files from virtual sources (such as other zip files)
	Network bool
}

var defaultConfig = Config{}
