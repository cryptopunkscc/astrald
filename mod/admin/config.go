package admin

const defaultPrompt = "> "

type Config struct {
	Prompt string `yaml:"prompt"`
	Admins []string
}

var defaultConfig = Config{
	Prompt: defaultPrompt,
}
