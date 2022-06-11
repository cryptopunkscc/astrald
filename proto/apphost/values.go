package apphost

const (
	success              = 0x00
	errRejected          = 0x01
	errFailed            = 0x02
	errTimeout           = 0x03
	errAlreadyRegistered = 0x04
	errUnexpected        = 0xff
)

const (
	cmdRegister = "register"
	cmdQuery    = "query"
	cmdResolve  = "resolve"
	cmdNodeInfo = "nodeInfo"
)
