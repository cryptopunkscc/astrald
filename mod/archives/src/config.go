package archives

type Config struct {
	AutoIndexZones string // list of zones (l - local, v - virtual, n - network)
}

var defaultConfig = Config{
	AutoIndexZones: "lv",
}
