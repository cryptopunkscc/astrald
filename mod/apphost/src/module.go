package apphost

import (
	"context"
	"errors"
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/log"
	"github.com/cryptopunkscc/astrald/debug"
	"github.com/cryptopunkscc/astrald/mod/admin"
	"github.com/cryptopunkscc/astrald/mod/apphost"
	"github.com/cryptopunkscc/astrald/mod/auth"
	"github.com/cryptopunkscc/astrald/mod/content"
	"github.com/cryptopunkscc/astrald/mod/dir"
	"github.com/cryptopunkscc/astrald/mod/objects"
	"github.com/cryptopunkscc/astrald/sig"
	"gorm.io/gorm"
	"net"
	"sync"
)

var _ apphost.Module = &Module{}

type Deps struct {
	Admin   admin.Module
	Auth    auth.Module
	Content content.Module
	Dir     dir.Module
	Objects objects.Module
}

type Module struct {
	Deps
	config Config
	node   astral.Node
	log    *log.Logger
	db     *gorm.DB

	listeners []net.Listener
	conns     <-chan net.Conn
	defaultID *astral.Identity
	guests    sig.Map[string, *Guest]
}

func (mod *Module) Run(ctx context.Context) error {
	var wg sync.WaitGroup
	var workerCount = mod.config.Workers

	mod.conns = mod.listen(ctx)

	// spawn workers
	mod.log.Logv(2, "spawning %v workers", workerCount)
	wg.Add(workerCount)
	for i := 0; i < workerCount; i++ {
		go func(i int) {
			defer debug.SaveLog(debug.SigInt)

			defer wg.Done()
			if err := mod.worker(ctx); err != nil {
				mod.log.Error("[%v] error: %v", i, err)
			}
		}(i)
	}

	// start the object server
	objectServer := NewObjectServer(mod)
	objectServer.Run(ctx)

	wg.Wait()

	return nil
}

func (mod *Module) SetDefaultIdentity(identity *astral.Identity) error {
	mod.defaultID = identity
	return nil
}

func (mod *Module) DefaultIdentity() *astral.Identity {
	return mod.defaultID
}

func (mod *Module) RegisterApp(appID string) (id *astral.Identity, err error) {
	err = mod.db.
		Model(&dbApp{}).
		Where("app_id = ?", appID).
		Select("identity").
		First(&id).Error

	if err == nil {
		return
	}

	id, err = astral.GenerateIdentity()
	if err != nil {
		return
	}

	err = mod.db.Create(&dbApp{
		AppID:    appID,
		Identity: id,
	}).Error

	return
}

func (mod *Module) UnregisterApp(appID string) (err error) {
	var found bool
	err = mod.db.
		Model(&dbApp{}).
		Where("app_id = ?", appID).
		Select("count(*)>0").
		First(&found).Error

	if err != nil {
		return err
	}
	if !found {
		return errors.New("app not found")
	}

	err = mod.db.Delete(&dbApp{AppID: appID}).Error

	return
}

func (mod *Module) ListApps() (list []string) {
	mod.db.
		Model(&dbApp{}).
		Select("app_id").
		Find(&list)
	return
}

func (mod *Module) AppToken(appID string) (token string, err error) {
	var row dbApp
	err = mod.db.Where("app_id = ?", appID).First(&row).Error
	if err != nil {
		return
	}

	return mod.FindOrCreateAccessToken(row.Identity)
}

func (mod *Module) String() string {
	return apphost.ModuleName
}
