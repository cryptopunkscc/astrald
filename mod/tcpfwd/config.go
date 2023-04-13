package tcpfwd

type Config struct {
	Out map[string]string `yaml:"out"`
	In  map[string]string `yaml:"in"`
}
