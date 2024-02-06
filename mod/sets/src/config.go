package sets

type configDisplay struct {
	Device   string
	Network  string
	Virtual  string
	Universe string
}

type Config struct {
	Display configDisplay
}

var defaultConfig = Config{
	Display: configDisplay{
		Device:   "Device",
		Network:  "Network",
		Virtual:  "Virtual",
		Universe: "Universe",
	},
}
