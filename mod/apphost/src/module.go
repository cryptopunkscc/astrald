package apphost

import (
	"context"
	"errors"
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/debug"
	"github.com/cryptopunkscc/astrald/log"
	"github.com/cryptopunkscc/astrald/mod/apphost"
	"github.com/cryptopunkscc/astrald/mod/content"
	"github.com/cryptopunkscc/astrald/mod/discovery"
	node2 "github.com/cryptopunkscc/astrald/node"
	"gorm.io/gorm"
	"net"
	"os"
	"path/filepath"
	"sync"
)

var _ apphost.Module = &Module{}

type Module struct {
	config  Config
	node    node2.Node
	content content.Module
	sdp     discovery.Module
	log     *log.Logger
	db      *gorm.DB

	listeners []net.Listener
	conns     <-chan net.Conn
	defaultID id.Identity
	guests    map[string]*Guest
	guestsMu  sync.Mutex
	execs     []*Exec
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
			identity, err := mod.node.Resolver().Resolve(run.Identity)
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

func (mod *Module) SetDefaultIdentity(identity id.Identity) error {
	mod.defaultID = identity
	return nil
}

func (mod *Module) DefaultIdentity() id.Identity {
	return mod.defaultID
}

func (mod *Module) addGuestRoute(identity id.Identity, name string, target string) error {
	mod.guestsMu.Lock()
	defer mod.guestsMu.Unlock()

	if len(name) == 0 {
		return errors.New("invalid name")
	}

	var key = identity.PublicKeyHex()

	var guest *Guest
	if g, found := mod.guests[key]; found {
		guest = g
	} else {
		guest = NewGuest(identity)
		mod.node.Router().AddRoute(id.Anyone, identity, guest, 90)
		mod.guests[key] = guest
	}

	relay := &RelayRouter{
		log:      mod.log,
		target:   target,
		identity: identity,
	}

	return guest.AddRoute(name, relay)
}

func (mod *Module) removeGuestRoute(identity id.Identity, name string) error {
	mod.guestsMu.Lock()
	defer mod.guestsMu.Unlock()

	var key = identity.PublicKeyHex()

	var guest = mod.guests[key]
	if guest == nil {
		return errors.New("route not found")
	}

	if err := guest.RemoveRoute(name); err != nil {
		return err
	}

	if guest.RouteCount() == 0 {
		delete(mod.guests, key)
		mod.node.Router().RemoveRoute(id.Anyone, identity, guest)
	}

	return nil
}

func (mod *Module) addNodeRoute(name string, target string) error {
	if len(name) == 0 {
		return errors.New("invalid name")
	}

	relay := &RelayRouter{
		log:      mod.log,
		target:   target,
		identity: mod.node.Identity(),
	}

	return mod.node.LocalRouter().AddRoute(name, relay)
}

func (mod *Module) removeNodeRoute(name string) error {
	return mod.node.LocalRouter().RemoveRoute(name)
}

func (mod *Module) getGuest(id id.Identity) *Guest {
	mod.guestsMu.Lock()
	defer mod.guestsMu.Unlock()

	var key = id.PublicKeyHex()

	return mod.guests[key]
}
