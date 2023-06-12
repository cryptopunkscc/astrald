package contacts

import (
	"context"
	"errors"
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/log"
	"github.com/cryptopunkscc/astrald/node"
	"github.com/cryptopunkscc/astrald/node/modules"
	"gorm.io/gorm"
)

var _ modules.Module = &Module{}

type Module struct {
	config Config
	node   node.Node
	log    *log.Logger
	db     *gorm.DB
}

func (m *Module) Run(ctx context.Context) error {
	if err := m.setupDatabase(); err != nil {
		return err
	}

	<-ctx.Done()
	return nil
}

func (m *Module) All() ([]Node, error) {
	var dbNodes []dbNode

	res := m.db.Find(&dbNodes)
	if res.Error != nil {
		return nil, res.Error
	}

	var nodes = make([]Node, 0, len(dbNodes))
	for _, n := range dbNodes {
		identity, err := id.ParsePublicKeyHex(n.Identity)
		if err != nil {
			continue
		}
		nodes = append(nodes, Node{
			Identity: identity,
			Alias:    n.Alias,
		})
	}

	return nodes, nil
}

func (m *Module) Find(identity id.Identity) (Node, error) {
	if identity.IsZero() {
		return Node{}, errors.New("zero identity")
	}
	keyHex := identity.PublicKeyHex()

	var dbNode dbNode

	if tx := m.db.First(&dbNode, "identity = ?", keyHex); tx.Error != nil {
		return Node{}, tx.Error
	}

	var node Node
	var err error

	node.Alias = dbNode.Alias
	node.Identity, err = id.ParsePublicKeyHex(dbNode.Identity)
	if err != nil {
		return Node{}, err
	}

	return node, nil
}

func (m *Module) FindOrCreate(identity id.Identity) (Node, error) {
	if node, err := m.Find(identity); err == nil {
		return node, nil
	}

	m.db.Create(&dbNode{
		Identity: identity.PublicKeyHex(),
	})

	return Node{
		Identity: identity,
	}, nil
}

func (m *Module) FindByAlias(alias string) (Node, error) {
	var dbNode dbNode

	if tx := m.db.First(&dbNode, "alias = ?", alias); tx.Error != nil {
		return Node{}, tx.Error
	}

	var node Node
	var err error

	node.Alias = dbNode.Alias
	node.Identity, err = id.ParsePublicKeyHex(dbNode.Identity)
	if err != nil {
		return Node{}, err
	}

	return node, nil
}

func (m *Module) Save(node Node) error {
	return m.db.Save(&dbNode{
		Identity: node.Identity.PublicKeyHex(),
		Alias:    node.Alias,
	}).Error
}

func (m *Module) Delete(identity id.Identity) error {
	return m.db.Delete(&dbNode{}, "identity = ?", identity.PublicKeyHex()).Error
}
