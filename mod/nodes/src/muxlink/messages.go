package muxlink

const (
	codeQuery = iota
	codeGrowBuffer
	codeReset
	codePing
	codePong
)

const (
	errSuccess = iota
	errRejected
	errRouteNotFound
	errUnexpected
)

type Header struct {
	Code int `cslq:"c"`
}

type Reset struct {
	Port int `cslq:"s"`
}

type GrowBuffer struct {
	Port int `cslq:"s"`
	Size int `cslq:"l"`
}

type Ping struct {
	Nonce int `cslq:"l"`
}

type Pong struct {
	Nonce int `cslq:"l"`
}

type Query struct {
	Query  string `cslq:"[c]c"`
	Port   int    `cslq:"s"`
	Buffer int    `cslq:"l"`
	Nonce  uint64 `cslq:"q"`
}

type Response struct {
	Error  int `cslq:"c"`
	Port   int `cslq:"s"`
	Buffer int `cslq:"l"`
}
