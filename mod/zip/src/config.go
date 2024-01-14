package zip

type Config struct {
	NoVirtual bool // don't automatically index zip files from virtual sources (such as other zip files)
}

var defaultConfig = Config{
	NoVirtual: true,
}
