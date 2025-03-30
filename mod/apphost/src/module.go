package apphost

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/log"
	"github.com/cryptopunkscc/astrald/debug"
	"github.com/cryptopunkscc/astrald/mod/admin"
	"github.com/cryptopunkscc/astrald/mod/apphost"
	"github.com/cryptopunkscc/astrald/mod/auth"
	"github.com/cryptopunkscc/astrald/mod/content"
	"github.com/cryptopunkscc/astrald/mod/dir"
	"github.com/cryptopunkscc/astrald/mod/objects"
	"github.com/cryptopunkscc/astrald/mod/shell"
	"github.com/cryptopunkscc/astrald/sig"
	"math/rand"
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
	db     *DB
	scope  shell.Scope

	listeners []net.Listener
	conns     <-chan net.Conn
	guests    sig.Map[string, *Guest]
}

var _ shell.HasScope = &Module{}

func (mod *Module) Run(ctx *astral.Context) error {
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

func (mod *Module) ListAccessTokens() ([]*apphost.AccessToken, error) {
	rows, err := mod.db.ListAccessTokens()
	if err != nil {
		return nil, err
	}

	return sig.MapSlice(rows, func(a dbAccessToken) (*apphost.AccessToken, error) {
		return &apphost.AccessToken{
			Identity:  a.Identity,
			Token:     astral.String8(a.Token),
			ExpiresAt: astral.Time(a.ExpiresAt),
		}, nil
	})
}

func (mod *Module) CreateAccessToken(identity *astral.Identity, d astral.Duration) (*apphost.AccessToken, error) {
	token, err := mod.db.CreateAccessToken(identity, d)
	if err != nil {
		return nil, err
	}

	return &apphost.AccessToken{
		Identity:  token.Identity,
		Token:     astral.String8(token.Token),
		ExpiresAt: astral.Time(token.ExpiresAt),
	}, nil
}

func (mod *Module) Scope() *shell.Scope {
	return &mod.scope
}

func (mod *Module) String() string {
	return apphost.ModuleName
}

func randomString(length int) (s string) {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789_"
	var name = make([]byte, length)
	for i := 0; i < len(name); i++ {
		name[i] = charset[rand.Intn(len(charset))]
	}
	return string(name[:])
}
