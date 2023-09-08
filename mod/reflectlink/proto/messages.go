package proto

type Endpoint struct {
	Network string `json:"network"`
	Address string `json:"address"`
}

type Reflection struct {
	RemoteEndpoint Endpoint `json:"endpoint,omitempty"`
}
