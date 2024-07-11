package nodes

import (
	"errors"
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/net"
	"gorm.io/gorm/clause"
)

func (mod *Module) Endpoints(nodeID id.Identity) (endpoints []net.Endpoint) {
	var rows []dbEndpoint

	err := mod.db.Find(&rows, "identity = ?", nodeID).Error
	if err != nil {
		return
	}

	for _, row := range rows {
		e, err := mod.node.Infra().Parse(row.Network, row.Address)
		if err != nil {
			mod.log.Errorv(1, "Endpoints(): error parsing db row: %v", err)
			continue
		}
		endpoints = append(endpoints, e)
	}

	return
}

func (mod *Module) AddEndpoint(nodeID id.Identity, endpoint ...net.Endpoint) error {
	var errs []error
	var err error
	for _, e := range endpoint {
		err = mod.addEndpoint(nodeID, e)
		if err != nil {
			errs = append(errs, err)
		}
	}
	return errors.Join(errs...)
}

func (mod *Module) addEndpoint(nodeID id.Identity, endpoint net.Endpoint) error {
	return mod.db.Clauses(clause.OnConflict{DoNothing: true}).
		Create(&dbEndpoint{
			Identity: nodeID,
			Network:  endpoint.Network(),
			Address:  endpoint.String(),
		}).Error
}

func (mod *Module) RemoveEndpoint(nodeID id.Identity, endpoint ...net.Endpoint) error {
	var errs []error
	var err error
	for _, e := range endpoint {
		err = mod.removeEndpoint(nodeID, e)
		if err != nil {
			errs = append(errs, err)
		}
	}
	return errors.Join(errs...)
}

func (mod *Module) removeEndpoint(nodeID id.Identity, endpoint net.Endpoint) error {
	return mod.db.Delete(&dbEndpoint{
		Identity: nodeID,
		Network:  endpoint.Network(),
		Address:  endpoint.String(),
	}).Error
}
