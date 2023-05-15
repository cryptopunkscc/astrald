package admin

const defaultPrompt = "> "

type Config struct {
	Prompt string `yaml:"prompt"`
}

var defaultConfig = Config{
	Prompt: defaultPrompt,
}
