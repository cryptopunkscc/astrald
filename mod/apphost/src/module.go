package apphost

import (
	"context"
	"errors"
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/debug"
	"github.com/cryptopunkscc/astrald/lib/routers"
	"github.com/cryptopunkscc/astrald/log"
	"github.com/cryptopunkscc/astrald/mod/admin"
	"github.com/cryptopunkscc/astrald/mod/apphost"
	"github.com/cryptopunkscc/astrald/mod/content"
	"github.com/cryptopunkscc/astrald/mod/dir"
	"gorm.io/gorm"
	"net"
	"os"
	"path/filepath"
	"sync"
)

var _ apphost.Module = &Module{}

type Deps struct {
	Admin   admin.Module
	Content content.Module
	Dir     dir.Module
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
	guests    map[string]*Guest
	guestsMu  sync.Mutex
	execs     []*Exec
	router    *routers.PathRouter
}

func (mod *Module) Run(ctx context.Context) error {
	var wg sync.WaitGroup
	var workerCount = mod.config.Workers

	mod.conns = mod.listen(ctx)

	mod.log.Infov(2, "running %d workers", workerCount)

	wg.Add(workerCount)
	for i := 0; i < workerCount; i++ {
		go func(i int) {
			defer debug.SaveLog(debug.SigInt)

			defer wg.Done()
			if err := mod.worker(ctx); err != nil {
				mod.log.Error("[%d] error: %s", i, err)
			}
		}(i)
	}

	if len(mod.config.Autorun) > 0 {
		mod.log.Infov(1, "%d autorun entries", len(mod.config.Autorun))
	}

	for _, run := range mod.config.Autorun {
		run := run
		go func() {
			identity, err := mod.Dir.Resolve(run.Identity)
			if err != nil {
				mod.log.Error("unknown identity: %s", run.Identity)
				return
			}

			var basename = filepath.Base(run.Exec)

			mod.log.Infov(1, "starting %s as %s...", basename, identity)

			exec, err := mod.Exec(identity, run.Exec, run.Args, os.Environ())
			if err != nil {
				mod.log.Errorv(0, "%s (%s) failed to start: %s", basename, identity, err)
				return
			}

			<-exec.Done()

			err = exec.err
			if err != nil {
				mod.log.Errorv(1, "%s (%s) exited with error: %s", basename, identity, err)
			}
		}()
	}

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

func (mod *Module) addGuestRoute(identity *astral.Identity, name string, target string) error {
	mod.guestsMu.Lock()
	defer mod.guestsMu.Unlock()

	if len(name) == 0 {
		return errors.New("invalid name")
	}

	var key = identity.String()

	var guest *Guest
	if g, found := mod.guests[key]; found {
		guest = g
	} else {
		guest = NewGuest(identity)
		mod.guests[key] = guest
	}

	relay := &RelayRouter{
		log:      mod.log,
		target:   target,
		identity: identity,
	}

	return guest.AddRoute(name, relay)
}

func (mod *Module) removeGuestRoute(identity *astral.Identity, name string) error {
	mod.guestsMu.Lock()
	defer mod.guestsMu.Unlock()

	var key = identity.String()

	var guest = mod.guests[key]
	if guest == nil {
		return errors.New("route not found")
	}

	if err := guest.RemoveRoute(name); err != nil {
		return err
	}

	if guest.RouteCount() == 0 {
		delete(mod.guests, key)
	}

	return nil
}

func (mod *Module) getGuest(id *astral.Identity) *Guest {
	mod.guestsMu.Lock()
	defer mod.guestsMu.Unlock()

	var key = id.String()

	return mod.guests[key]
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
