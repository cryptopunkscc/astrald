package ip

const ModuleName = "Ip"

type Module interface {
	LocalIPs() (ips []IP, err error)
	PublicIPs() (ips []IP)
}
