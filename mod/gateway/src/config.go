package gateway

const defaultGateway = "node1f3AwbE1gJAgAqEx98FMipokcaE9ZapIphzDUkAceE7Pmw8ghmFV19QKCATeC7uyoLszQA"

type Config struct {
	Subscribe []string `yaml:"subscribe"`
}

var defaultConfig = Config{
	Subscribe: []string{
		defaultGateway,
	},
}
