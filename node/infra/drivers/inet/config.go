package inet

const (
	defaultListenPort = 1791
)

type Config struct {
	PublicAddr []string `yaml:"public_addr"`
	ListenPort int      `yaml:"listen_port"`
}

var defaultConfig = Config{
	ListenPort: defaultListenPort,
}
