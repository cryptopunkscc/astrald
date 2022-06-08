package apphost

const (
	success     = 0x00
	errRejected = 0x01
	errFailed   = 0x02
)

const (
	cmdRegister = "register"
	cmdQuery    = "query"
	cmdResolve  = "resolve"
	cmdNodeInfo = "nodeInfo"
)
