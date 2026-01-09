package apphost

import (
	"errors"
	"fmt"
	"math/rand"
	"net"
	"sync"
	"time"

	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/log"
	"github.com/cryptopunkscc/astrald/debug"
	"github.com/cryptopunkscc/astrald/mod/apphost"
	"github.com/cryptopunkscc/astrald/mod/auth"
	"github.com/cryptopunkscc/astrald/mod/dir"
	"github.com/cryptopunkscc/astrald/mod/keys"
	"github.com/cryptopunkscc/astrald/mod/objects"
	"github.com/cryptopunkscc/astrald/mod/shell"
	"github.com/cryptopunkscc/astrald/sig"
)

var _ apphost.Module = &Module{}

type Deps struct {
	Auth    auth.Module
	Dir     dir.Module
	Keys    keys.Module
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
	handlers  sig.Set[*QueryHandler]
	enRoute   sig.Map[astral.Nonce, *queryEnRoute]
	indexMu   sync.Mutex
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
				mod.log.Error("[worker:%v] error: %v", i, err)
			}
		}(i)
	}

	// start the indexer
	go mod.indexer(ctx)

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

func (mod *Module) SignAppContract(c *apphost.AppContract) (err error) {
	hash := c.ContractHash()

	// add app signature
	c.AppSig, err = mod.Keys.SignASN1(c.AppID, hash)
	if err != nil {
		return
	}

	// add host signature
	c.HostSig, err = mod.Keys.SignASN1(c.HostID, hash)

	return
}

func (mod *Module) ActiveLocalAppContracts() (list []*apphost.AppContract, err error) {
	contracts, err := mod.db.FindActiveAppContractsByHost(mod.node.Identity())
	if err != nil {
		return
	}

	for _, dbContract := range contracts {
		contract, err := objects.Load[*apphost.AppContract](nil, mod.Objects.ReadDefault(), dbContract.ObjectID)
		if err != nil {
			mod.log.Errorv(2, "error loading contract %v: %v", dbContract.ObjectID, err)
			continue
		}
		list = append(list, contract)
	}

	return
}

func (mod *Module) Index(ctx *astral.Context, objectID *astral.ObjectID) (err error) {
	mod.indexMu.Lock()
	defer mod.indexMu.Unlock()

	// check if already indexed
	if _, e := mod.db.FindAppContract(objectID); e == nil {
		return nil
	}

	// load the contract from node repo
	c, err := objects.Load[*apphost.AppContract](ctx, mod.Objects.ReadDefault(), objectID)
	if err != nil {
		return fmt.Errorf("cannot load app contract: %w", err)
	}

	if !mod.isActive(c) {
		return errors.New("inactive contract")
	}

	// save the contract
	err = mod.db.SaveAppContract(&dbAppContract{
		ObjectID:  objectID,
		AppID:     c.AppID,
		HostID:    c.HostID,
		StartsAt:  time.Time(c.StartsAt),
		ExpiresAt: time.Time(c.ExpiresAt),
	})
	if err != nil {
		return err
	}

	mod.Objects.Receive(&apphost.EventNewAppContract{Contract: c}, nil)

	return
}

// isActive returns true if the contract is active, i.e., the signatures are valid, and its conditions are met (such
// as start and expiry date).
func (mod *Module) isActive(c *apphost.AppContract) bool {
	switch {
	case c.StartsAt.Time().After(time.Now()):
		return false // hasn't started yet

	case c.ExpiresAt.Time().Before(time.Now()):
		return false // has expired
	case mod.validateSignatures(c) != nil:
		return false // invalid signatures
	}

	return true
}

func (mod *Module) validateSignatures(c *apphost.AppContract) (err error) {
	hash := c.ContractHash()

	// verify app signature
	err = mod.Keys.VerifyASN1(c.AppID, hash, c.AppSig)
	if err != nil {
		return
	}

	// verify host signature
	err = mod.Keys.VerifyASN1(c.HostID, hash, c.HostSig)

	return
}

func (mod *Module) indexer(ctx *astral.Context) {
	ctx = ctx.ExcludeZone(astral.ZoneNetwork)

	ch, err := mod.Objects.GetRepository(objects.RepoLocal).Scan(ctx, true)
	if err != nil {
		mod.log.Error("cannot scan objects: %v", err)
		return
	}

	for objectID := range ch {
		objectType, err := mod.Objects.GetType(ctx, objectID)

		switch {
		case err != nil:
			continue
		case objectType != apphost.AppContract{}.ObjectType():
			continue
		}

		_ = mod.Index(ctx, objectID)
	}

	mod.log.Logv(1, "apphost indexer finished")
}

func randomString(length int) (s string) {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789_"
	var name = make([]byte, length)
	for i := 0; i < len(name); i++ {
		name[i] = charset[rand.Intn(len(charset))]
	}
	return string(name[:])
}
