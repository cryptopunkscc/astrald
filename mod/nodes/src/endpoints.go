package nodes

import (
	"errors"
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/mod/exonet"
	"gorm.io/gorm/clause"
)

func (mod *Module) Endpoints(nodeID *astral.Identity) (endpoints []exonet.Endpoint) {
	endpoints, _ = mod.Exonet.ResolveEndpoints(mod.ctx, nodeID)
	return
}

func (mod *Module) hasEndpoints(nodeID *astral.Identity) (has bool) {
	return len(mod.Endpoints(nodeID)) > 0
}

func (mod *Module) AddEndpoint(nodeID *astral.Identity, endpoint ...exonet.Endpoint) error {
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

func (mod *Module) addEndpoint(nodeID *astral.Identity, endpoint exonet.Endpoint) error {
	return mod.db.Clauses(clause.OnConflict{DoNothing: true}).
		Create(&dbEndpoint{
			Identity: nodeID,
			Network:  endpoint.Network(),
			Address:  endpoint.Address(),
		}).Error
}

func (mod *Module) RemoveEndpoint(nodeID *astral.Identity, endpoint ...exonet.Endpoint) error {
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

func (mod *Module) removeEndpoint(nodeID *astral.Identity, endpoint exonet.Endpoint) error {
	return mod.db.Delete(&dbEndpoint{
		Identity: nodeID,
		Network:  endpoint.Network(),
		Address:  endpoint.Address(),
	}).Error
}
