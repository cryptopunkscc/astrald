package apphost

import (
	"context"
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/log"
	"github.com/cryptopunkscc/astrald/mod/contacts"
	"github.com/cryptopunkscc/astrald/node"
	"github.com/cryptopunkscc/astrald/node/modules"
	"math/rand"
	"net"
	"os"
	"sync"
)

type Module struct {
	config    Config
	node      node.Node
	conns     <-chan net.Conn
	log       *log.Logger
	listeners []net.Listener
	tokens    map[string]id.Identity
	mu        sync.Mutex
}

func (mod *Module) Run(ctx context.Context) error {
	var wg sync.WaitGroup
	var workerCount = mod.config.Workers

	mod.conns = mod.listen(ctx)

	mod.log.Infov(2, "running %d workers", workerCount)

	wg.Add(workerCount)
	for i := 0; i < workerCount; i++ {
		go func(i int) {
			defer wg.Done()
			if err := mod.worker(ctx); err != nil {
				mod.log.Error("[%d] error: %s", i, err)
			}
		}(i)
	}

	if len(mod.config.Boot) > 0 {
		mod.log.Infov(1, "starting boot apps...")
	}

	_, err := modules.WaitReady[*contacts.Module](ctx, mod.node.Modules())
	if err != nil {
		mod.log.Errorv(1, "error waiting for mod_contacts: %s", err)
	}

	for _, boot := range mod.config.Boot {
		boot := boot
		go func() {
			mod.log.Infov(1, "starting %s...", boot.App)
			err := mod.Launch(boot.App, boot.Args, os.Environ())
			if err != nil {
				mod.log.Infov(1, "app %s ended with error: %s", boot.App, err)
			}
		}()
	}

	wg.Wait()

	return nil
}

func (mod *Module) authToken(token string) id.Identity {
	mod.mu.Lock()
	defer mod.mu.Unlock()

	if s, ok := mod.config.Tokens[token]; ok {
		if i, err := mod.node.Resolver().Resolve(s); err == nil {
			return i
		}
	}

	if identity, found := mod.tokens[token]; found {
		return identity
	}

	return id.Identity{}
}

func (mod *Module) createToken(identity id.Identity) string {
	mod.mu.Lock()
	defer mod.mu.Unlock()

	var token = randomString(32)

	mod.tokens[token] = identity

	return token
}

func randomString(length int) (s string) {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789_"
	var name = make([]byte, length)
	for i := 0; i < len(name); i++ {
		name[i] = charset[rand.Intn(len(charset))]
	}
	return string(name[:])
}
