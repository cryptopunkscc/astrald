package proto

const (
	StatusError = "error"
	StatusOK    = "ok"
)

const (
	RequestQuery    = "connect"
	RequestRegister = "register"
)

type Request struct {
	Type     string
	Identity string
	Port     string
	Path     string
}

type Response struct {
	Status string
	Error  string
}
