package relay

import (
	"context"
	"errors"
	"fmt"
	"github.com/cryptopunkscc/astrald/core"
	"github.com/cryptopunkscc/astrald/core/assets"
	"github.com/cryptopunkscc/astrald/cslq"
	"github.com/cryptopunkscc/astrald/id"
	"github.com/cryptopunkscc/astrald/lib/adc"
	"github.com/cryptopunkscc/astrald/log"
	"github.com/cryptopunkscc/astrald/mod/content"
	"github.com/cryptopunkscc/astrald/mod/keys"
	"github.com/cryptopunkscc/astrald/mod/objects"
	"github.com/cryptopunkscc/astrald/mod/relay"
	"github.com/cryptopunkscc/astrald/net"
	node2 "github.com/cryptopunkscc/astrald/node"
	"github.com/cryptopunkscc/astrald/object"
	"github.com/cryptopunkscc/astrald/streams"
	"github.com/cryptopunkscc/astrald/tasks"
	"gorm.io/gorm"
	"sync"
)

var _ relay.Module = &Module{}

type Module struct {
	node     node2.Node
	assets   assets.Assets
	log      *log.Logger
	config   Config
	ctx      context.Context
	routes   map[string]id.Identity
	routesMu sync.Mutex
	db       *gorm.DB
	objects  objects.Module
	content  content.Module
	keys     keys.Module
}

func (mod *Module) Run(ctx context.Context) error {
	mod.ctx = ctx

	return tasks.Group(
		&IndexerService{Module: mod},
		&RelayService{Module: mod},
		&RerouteService{Module: mod},
	).Run(ctx)
}

func (mod *Module) Save(cert *relay.Cert) (objectID object.ID, err error) {
	// validate the certificate
	err = cert.Validate()
	if err != nil {
		return
	}

	// create storage writer
	w, err := mod.objects.Create(nil)
	if err != nil {
		return
	}

	// encode the header and the cert
	err = cslq.Encode(w, "vv", adc.Header(relay.CertType), cert)
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

func (mod *Module) Reroute(nonce net.Nonce, router net.Router) error {
	conn := mod.findConnByNonce(nonce)
	if conn == nil {
		return errors.New("conn not found")
	}

	routerIdentity := mod.getRouter(conn.Target())
	if routerIdentity.IsZero() {
		return errors.New("cannot establish router identity")
	}

	serviceQuery := net.NewQuery(mod.node.Identity(), routerIdentity, relay.RerouteServiceName)
	serviceConn, err := net.Route(mod.ctx, router, serviceQuery)
	if err != nil {
		return err
	}

	if err := cslq.Encode(serviceConn, "q", nonce); err != nil {
		return err
	}

	var errCode int
	cslq.Decode(serviceConn, "c", &errCode)
	if errCode != 0 {
		return fmt.Errorf("error code %d", errCode)
	}

	switcher, err := mod.insertSwitcherAfter(net.RootSource(conn.Caller()))
	if err != nil {
		return err
	}

	newRoot, ok := net.RootSource(serviceConn).(net.OutputGetSetter)
	if !ok {
		return errors.New("newroot not an OutputGetSetter")
	}

	debris := newRoot.Output()
	newRoot.SetOutput(switcher.NextWriter)

	newOutput := mod.yankFinalOutput(serviceConn)
	oldOutput := net.FinalOutput(conn.Target())
	if err := mod.replaceOutput(oldOutput, newOutput); err != nil {
		return err
	}

	switcher.AfterSwitch = func() {
		debris.Close()
		serviceConn.Close()
	}

	return oldOutput.Close()
}

func (mod *Module) yankFinalOutput(chain any) net.SecureWriteCloser {
	final := net.FinalOutput(chain)

	s, ok := final.(net.SourceGetSetter)
	if !ok {
		return nil
	}

	prev, ok := s.Source().(net.OutputGetSetter)
	if !ok {
		return nil
	}

	prev.SetOutput(net.NewSecurePipeWriter(streams.NilWriteCloser{}, id.Identity{}))
	s.SetSource(nil)

	return final
}

func (mod *Module) replaceOutput(old, new net.SecureWriteCloser) error {
	var prev net.OutputSetter

	if old == nil {
		panic("old is nil")
	}
	if new == nil {
		panic("new is nil")
	}

	s, ok := old.(net.SourceGetter)
	if !ok {
		return errors.New("old output is not a SourceGetter")
	}

	prev, ok = s.Source().(net.OutputSetter)
	if !ok {
		return errors.New("source is not an OutputSetter")
	}

	return prev.SetOutput(new)
}

func (mod *Module) insertSwitcherAfter(item any) (*SwitchWriter, error) {
	i, ok := item.(net.OutputGetSetter)
	if !ok {
		return nil, fmt.Errorf("argument not an OutputGetSetter")
	}

	switcher := NewSwitchWriter(i.Output())
	i.SetOutput(switcher)
	if s, ok := switcher.Output().(net.SourceSetter); ok {
		s.SetSource(switcher)
	}

	return switcher, nil
}

func (mod *Module) findConnByNonce(nonce net.Nonce) *core.MonitoredConn {
	coreRouter, ok := mod.node.Router().(*core.CoreRouter)
	if !ok {
		return nil
	}

	for _, c := range coreRouter.Conns().All() {
		if c.Query().Nonce() == nonce {
			return c
		}
	}
	return nil
}

func (mod *Module) isLocal(identity id.Identity) bool {
	return mod.node.Identity().IsEqual(identity)
}

func (mod *Module) getRouter(w net.SecureWriteCloser) id.Identity {
	if final := net.FinalOutput(w); final != nil {
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

	r, err := mod.objects.Open(context.Background(), objectID, objects.DefaultOpenOpts())
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
