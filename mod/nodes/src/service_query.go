package nodes

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/mod/nodes"
	"gorm.io/gorm"
	"time"
)

type ServiceQuery struct {
	db *gorm.DB
}

func (s *ServiceQuery) Find() (list []*nodes.Service) {
	var rows []*dbService
	err := s.db.
		Where("expires_at > ?", time.Now().UTC()).
		Find(&rows).Error
	if err != nil {
		return
	}

	for _, row := range rows {
		list = append(list, &nodes.Service{
			ProviderID: row.ProviderID,
			Name:       astral.String8(row.Name),
			Priority:   astral.Uint16(row.Priority),
		})
	}

	return
}

func (s *ServiceQuery) ByName(name ...string) nodes.ServiceQuery {
	return &ServiceQuery{
		db: s.db.Where("name IN (?)", name),
	}
}

func (s *ServiceQuery) ByNodeID(id ...*astral.Identity) nodes.ServiceQuery {
	return &ServiceQuery{
		db: s.db.Where("provider_id IN (?)", id),
	}
}
