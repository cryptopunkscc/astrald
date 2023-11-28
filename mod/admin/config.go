package admin

const defaultPrompt = "> "
const timestampFormat string = "2006-01-02 15:04:05"

type Config struct {
	Prompt string `yaml:"prompt"`
	Admins []string
}

var defaultConfig = Config{
	Prompt: defaultPrompt,
}
