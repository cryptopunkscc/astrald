package astral

import (
	"github.com/cryptopunkscc/astrald/cslq"
	"github.com/cryptopunkscc/astrald/id"
	"github.com/cryptopunkscc/astrald/mod/discovery"
	"github.com/cryptopunkscc/astrald/mod/discovery/proto"
)

type DiscoveryHandler func(remoteID id.Identity) []ServiceInfo

type Discovery struct {
	*ApphostClient
}

type ServiceInfo struct {
	Identity id.Identity
	Name     string
	Type     string
	Extra    []byte
}

func NewDiscovery(apphost *ApphostClient) *Discovery {
	return &Discovery{ApphostClient: apphost}
}

func (d *Discovery) Discover(identity id.Identity) ([]ServiceInfo, error) {
	c, err := d.Query(identity, discovery.DiscoverServiceName)
	if err != nil {
		return nil, err
	}

	var list []ServiceInfo

	for err == nil {
		err = cslq.Invoke(c, func(msg proto.Service) error {
			list = append(list, ServiceInfo{
				Identity: msg.Identity,
				Name:     msg.Name,
				Type:     msg.Type,
				Extra:    msg.Extra,
			})
			return nil
		})
	}

	return list, nil
}
