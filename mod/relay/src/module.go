package relay

import (
	"context"
	"errors"
	"fmt"
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/core/assets"
	"github.com/cryptopunkscc/astrald/cslq"
	"github.com/cryptopunkscc/astrald/id"
	"github.com/cryptopunkscc/astrald/lib/routers"
	"github.com/cryptopunkscc/astrald/log"
	"github.com/cryptopunkscc/astrald/mod/admin"
	"github.com/cryptopunkscc/astrald/mod/content"
	"github.com/cryptopunkscc/astrald/mod/dir"
	"github.com/cryptopunkscc/astrald/mod/keys"
	"github.com/cryptopunkscc/astrald/mod/objects"
	"github.com/cryptopunkscc/astrald/mod/relay"
	"github.com/cryptopunkscc/astrald/object"
	"github.com/cryptopunkscc/astrald/streams"
	"github.com/cryptopunkscc/astrald/tasks"
	"gorm.io/gorm"
	"sync"
)

var _ relay.Module = &Module{}

type Deps struct {
	Admin   admin.Module
	Content content.Module
	Dir     dir.Module
	Keys    keys.Module
	Objects objects.Module
}

type Module struct {
	Deps
	*routers.PathRouter
	node     astral.Node
	assets   assets.Assets
	log      *log.Logger
	config   Config
	ctx      context.Context
	routes   map[string]id.Identity
	routesMu sync.Mutex
	db       *gorm.DB
}

func (mod *Module) Run(ctx context.Context) error {
	mod.ctx = ctx

	return tasks.Group(
		&IndexerService{Module: mod},
		&RelayService{Module: mod},
	).Run(ctx)
}

func (mod *Module) Save(cert *relay.Cert) (objectID object.ID, err error) {
	// validate the certificate
	err = cert.Validate()
	if err != nil {
		return
	}

	// create storage writer
	w, err := mod.Objects.Create(nil)
	if err != nil {
		return
	}

	// encode the header and the cert
	err = cslq.Encode(w, "vv", astral.ObjectHeader(relay.CertType), cert)
	if err != nil {
		return
	}

	// commit the object
	objectID, err = w.Commit()
	if err != nil {
		return
	}

	err = mod.index(cert)
	return
}

func (mod *Module) yankFinalOutput(chain any) astral.SecureWriteCloser {
	final := astral.FinalOutput(chain)

	s, ok := final.(astral.SourceGetSetter)
	if !ok {
		return nil
	}

	prev, ok := s.Source().(astral.OutputGetSetter)
	if !ok {
		return nil
	}

	prev.SetOutput(astral.NewSecurePipeWriter(streams.NilWriteCloser{}, id.Identity{}))
	s.SetSource(nil)

	return final
}

func (mod *Module) replaceOutput(old, new astral.SecureWriteCloser) error {
	var prev astral.OutputSetter

	if old == nil {
		panic("old is nil")
	}
	if new == nil {
		panic("new is nil")
	}

	s, ok := old.(astral.SourceGetter)
	if !ok {
		return errors.New("old output is not a SourceGetter")
	}

	prev, ok = s.Source().(astral.OutputSetter)
	if !ok {
		return errors.New("source is not an OutputSetter")
	}

	return prev.SetOutput(new)
}

func (mod *Module) insertSwitcherAfter(item any) (*SwitchWriter, error) {
	i, ok := item.(astral.OutputGetSetter)
	if !ok {
		return nil, fmt.Errorf("argument not an OutputGetSetter")
	}

	switcher := NewSwitchWriter(i.Output())
	i.SetOutput(switcher)
	if s, ok := switcher.Output().(astral.SourceSetter); ok {
		s.SetSource(switcher)
	}

	return switcher, nil
}

func (mod *Module) isLocal(identity id.Identity) bool {
	return mod.node.Identity().IsEqual(identity)
}

func (mod *Module) getRouter(w astral.SecureWriteCloser) id.Identity {
	if final := astral.FinalOutput(w); final != nil {
		return final.Identity()
	}
	return id.Identity{}
}

func (mod *Module) verifyIndex(objectID object.ID) error {
	var row dbCert
	var err = mod.db.
		Where("data_id = ?", objectID).
		First(&row).Error
	if err != nil {
		return nil
	}

	r, err := mod.Objects.Open(context.Background(), objectID, objects.DefaultOpenOpts())
	if err == nil {
		r.Close()
		return nil
	}

	err = mod.db.Delete(&row).Error
	if err != nil {
		mod.log.Errorv(2, "db: delete error: %v", err)
	}
	return nil
}
