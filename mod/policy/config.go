package policy

type Config struct {
	AlwaysLinked  *ConfigAlwaysLinked  `yaml:"always_linked,omitempty"`
	OptimizeLinks *ConfigOptimizeLinks `yaml:"optimize_links,omitempty"`
	RerouteConns  *ConfigRerouteConns  `yaml:"reroute_conns,omitempty"`
}

type ConfigAlwaysLinked struct {
	Targets []string `yaml:"targets"`
}

type ConfigOptimizeLinks struct {
}

type ConfigRerouteConns struct {
}

var defaultConfig = Config{
	AlwaysLinked:  &ConfigAlwaysLinked{},
	OptimizeLinks: &ConfigOptimizeLinks{},
	RerouteConns:  &ConfigRerouteConns{},
}
