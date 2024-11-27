package user

const (
	methodClaim = "user.claim"
	methodNodes = "user.nodes"
)

type Config struct {
	Identity string `yaml:"identity"`
}

var defaultConfig = Config{}
