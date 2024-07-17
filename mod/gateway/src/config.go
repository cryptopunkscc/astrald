package gateway

const defaultGateway = "node1f3AwbE1gB4AACoSE3zXwImiSypR0nplikGOPQRCw5J2fYCzDGaWUV3DpIAM5F2dlRXYnhA"

type Config struct {
	Subscribe []string `yaml:"subscribe"`
}

var defaultConfig = Config{
	Subscribe: []string{
		//defaultGateway,
	},
}
