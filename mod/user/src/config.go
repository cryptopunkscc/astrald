package user

const (
	methodClaim = "user.claim"
	methodNodes = "user.nodes"
)

type Config struct {
	Identity string `yaml:"identity"`
	Public   bool   `yaml:"public"`
}

var defaultConfig = Config{
	Public: true,
}
