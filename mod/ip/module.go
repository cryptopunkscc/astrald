package ip

const ModuleName = "ip"

type Module interface {
	LocalIPs() (ips []IP, err error)
	PublicIPs() (ips []IP)
}
