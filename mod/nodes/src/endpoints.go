package nodes

import (
	"errors"
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/id"
	"github.com/cryptopunkscc/astrald/mod/exonet"
	"github.com/cryptopunkscc/astrald/mod/nodes"
	"gorm.io/gorm/clause"
)

func (mod *Module) Endpoints(nodeID id.Identity) (endpoints []exonet.Endpoint) {
	var rows []dbEndpoint

	err := mod.db.Find(&rows, "identity = ?", nodeID).Error
	if err != nil {
		return
	}

	for _, row := range rows {
		e, err := mod.exonet.Parse(row.Network, row.Address)
		if err != nil {
			mod.log.Errorv(1, "Endpoints(): error parsing db row: %v", err)
			continue
		}
		endpoints = append(endpoints, e)
	}

	return
}

func (mod *Module) hasEndpoints(nodeID id.Identity) (has bool) {
	mod.db.
		Model(&dbEndpoint{}).
		Where("identity = ?", nodeID).
		Select("count(*) > 0").
		First(&has)
	return
}

func (mod *Module) AddEndpoint(nodeID id.Identity, endpoint ...exonet.Endpoint) error {
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

func (mod *Module) addEndpoint(nodeID id.Identity, endpoint exonet.Endpoint) error {
	return mod.db.Clauses(clause.OnConflict{DoNothing: true}).
		Create(&dbEndpoint{
			Identity: nodeID,
			Network:  endpoint.Network(),
			Address:  endpoint.Address(),
		}).Error
}

func (mod *Module) RemoveEndpoint(nodeID id.Identity, endpoint ...exonet.Endpoint) error {
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

func (mod *Module) removeEndpoint(nodeID id.Identity, endpoint exonet.Endpoint) error {
	return mod.db.Delete(&dbEndpoint{
		Identity: nodeID,
		Network:  endpoint.Network(),
		Address:  endpoint.Address(),
	}).Error
}

// Network returns link's network name or unknown if network could not be determined
func Network(link nodes.Link) string {
	type Transporter interface {
		Transport() astral.Conn
	}

	if l, ok := link.(Transporter); ok {
		var t = l.Transport().(exonet.Conn)
		if t == nil {
			return "unknown"
		}

		if e := t.RemoteEndpoint(); e != nil {
			return e.Network()
		}
		if e := t.LocalEndpoint(); e != nil {
			return e.Network()
		}
	}

	return "unknown"
}
