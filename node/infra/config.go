package infra

const configName = "infra"

type Config struct {
	Drivers []string `yaml:"drivers"`
}

var defaultConfig = Config{
	Drivers: []string{"inet", "gw"},
}

func (cfg Config) driversContain(driver string) bool {
	for _, d := range cfg.Drivers {
		if d == driver {
			return true
		}
	}
	return false
}
