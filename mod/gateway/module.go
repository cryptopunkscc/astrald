package gateway

const ModuleName = "gateway"

const (
	MethodBind    = "gateway.node_bind"
	MethodConnect = "gateway.node_connect"
	MethodList    = "gateway.node_list"
	MethodRoute   = "gateway.route"
)

type Module interface {
}
