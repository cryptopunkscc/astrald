package inet

const (
	defaultListenPort = 1791
)

type Config struct {
	PublicAddr        []string `yaml:"public_addr"`
	AnnounceOnlyIface string   `yaml:"announce_only_iface"`
	ListenPort        int      `yaml:"listen_port"`
}
