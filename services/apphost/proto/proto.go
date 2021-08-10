package proto

const (
	StatusError = "error"
	StatusOK    = "ok"
)

const (
	RequestConnect  = "connect"
	RequestRegister = "register"
)

type Request struct {
	Type     string `json:"type"`
	Identity string `json:"identity"`
	Port     string `json:"port"`
	Path     string `json:"path"`
}

type Response struct {
	Status string `json:"status"`
	Error  string `json:"error"`
}

type Decoder interface {
	Decode(e interface{}) error
}

type Encoder interface {
	Encode(e interface{}) error
}
