package apphost

import (
	"context"
	"errors"
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/debug"
	"github.com/cryptopunkscc/astrald/log"
	"github.com/cryptopunkscc/astrald/mod/admin"
	"github.com/cryptopunkscc/astrald/mod/sdp"
	"github.com/cryptopunkscc/astrald/node"
	"github.com/cryptopunkscc/astrald/node/assets"
	"github.com/cryptopunkscc/astrald/node/modules"
	"math/rand"
	"net"
	"os"
	"path/filepath"
	"sync"
)

type Module struct {
	config    Config
	node      node.Node
	keys      assets.KeyStore
	conns     <-chan net.Conn
	log       *log.Logger
	listeners []net.Listener
	tokens    map[string]id.Identity
	tokensMu  sync.Mutex
	guests    map[string]*Guest
	guestsMu  sync.Mutex
	execs     []*Exec
}

func (mod *Module) Run(ctx context.Context) error {
	// inject admin command
	if adm, err := modules.Find[*admin.Module](mod.node.Modules()); err == nil {
		_ = adm.AddCommand("apphost", &Admin{mod: mod})
	}

	if disco, err := modules.Find[*sdp.Module](mod.node.Modules()); err == nil {
		disco.AddSource(mod)
		defer disco.RemoveSource(mod)
	}

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

func (mod *Module) getGuest(id id.Identity) *Guest {
	mod.guestsMu.Lock()
	defer mod.guestsMu.Unlock()

	var key = id.PublicKeyHex()

	return mod.guests[key]
}

func (mod *Module) authToken(token string) (identity id.Identity) {
	mod.tokensMu.Lock()
	defer mod.tokensMu.Unlock()

	var err error

	if s, ok := mod.config.Tokens[token]; ok {
		identity, err = mod.node.Resolver().Resolve(s)
	}

	if identity.IsZero() {
		identity, _ = mod.tokens[token]
	}

	if identity.IsZero() {
		return identity
	}

	identity, err = mod.keys.Find(identity)
	if err != nil {
		return id.Identity{}
	}

	return identity
}

func (mod *Module) createToken(identity id.Identity) string {
	mod.tokensMu.Lock()
	defer mod.tokensMu.Unlock()

	var token = randomString(32)

	mod.tokens[token] = identity

	return token
}

func (mod *Module) defaultIdentity() id.Identity {
	i, _ := mod.node.Resolver().Resolve(mod.config.DefaultIdentity)
	return i
}

func randomString(length int) (s string) {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789_"
	var name = make([]byte, length)
	for i := 0; i < len(name); i++ {
		name[i] = charset[rand.Intn(len(charset))]
	}
	return string(name[:])
}
