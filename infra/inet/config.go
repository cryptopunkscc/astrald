package inet

import "time"

const (
	defaultListenPort   = 1791
	defaultPresencePort = 8829
	presencePayloadLen  = 36
	presenceInterval    = time.Minute
)

type Config struct {
	PublicAddr        []string `yaml:"public_addr"`
	AnnounceOnlyIface string   `yaml:"announce_only_iface"`
}
