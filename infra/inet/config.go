package inet

type Config struct {
	PublicAddr        []string `yaml:"public_addr"`
	AnnounceOnlyIface string   `yaml:"announce_only_iface"`
}
