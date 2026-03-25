package gateway

const ModuleName = "gateway"

const (
	MethodNodeRegister   = "gateway.node_register"
	MethodNodeUnregister = "gateway.node_unregister"
	MethodNodeConnect    = "gateway.node_connect"
	MethodNodeList       = "gateway.node_list"
	MethodNodeRoute      = "gateway.node_route"
)

type Module interface {
}
