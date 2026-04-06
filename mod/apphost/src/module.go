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
	"github.com/cryptopunkscc/astrald/lib/ops"
	"github.com/cryptopunkscc/astrald/mod/apphost"
	"github.com/cryptopunkscc/astrald/mod/auth"
	"github.com/cryptopunkscc/astrald/mod/crypto"
	"github.com/cryptopunkscc/astrald/mod/dir"
	"github.com/cryptopunkscc/astrald/mod/objects"
	"github.com/cryptopunkscc/astrald/mod/secp256k1"
	"github.com/cryptopunkscc/astrald/sig"
)

var _ apphost.Module = &Module{}

type Deps struct {
	Auth    auth.Module
	Crypto  crypto.Module
	Dir     dir.Module
	Objects objects.Module
}

type Module struct {
	Deps
	config Config
	node   astral.Node
	log    *log.Logger
	db     *DB
	scope  ops.Set

	listeners []net.Listener
	conns     <-chan net.Conn
	handlers  sig.Set[*QueryHandler]
	enRoute   sig.Map[astral.Nonce, *queryEnRoute]
	indexMu   sync.Mutex
}

var _ ops.HasOps = &Module{}

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
	objectServer := NewHTTPServer(mod)
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

func (mod *Module) AuthenticateToken(token string) (*astral.Identity, error) {
	dbToken, err := mod.db.FindAccessToken(token)
	if err != nil || dbToken == nil {
		return nil, errors.New("invalid token")
	}

	return dbToken.Identity, nil
}

func (mod *Module) GetOpSet() *ops.Set {
	return &mod.scope
}

func (mod *Module) String() string {
	return apphost.ModuleName
}

func (mod *Module) SignAppContract(ctx *astral.Context, c *apphost.AppContract) (*apphost.SignedAppContract, error) {
	signed := &apphost.SignedAppContract{AppContract: c}

	// sign as host (node) — ASN1 hash-based
	hostSigner, err := mod.Crypto.ObjectSigner(secp256k1.FromIdentity(c.HostID))
	if err != nil {
		return nil, fmt.Errorf("sign as host: %w", err)
	}
	signed.HostSig, err = hostSigner.SignObject(ctx, c)
	if err != nil {
		return nil, fmt.Errorf("sign as host: %w", err)
	}

	// sign as app — ASN1 hash-based
	appSigner, err := mod.Crypto.ObjectSigner(secp256k1.FromIdentity(c.AppID))
	if err != nil {
		return nil, fmt.Errorf("sign as app: %w", err)
	}
	signed.AppSig, err = appSigner.SignObject(ctx, c)
	if err != nil {
		return nil, fmt.Errorf("sign as app: %w", err)
	}

	return signed, nil
}

func (mod *Module) ActiveLocalAppContracts() (list []*apphost.SignedAppContract, err error) {
	contracts, err := mod.db.FindActiveAppContractsByHost(mod.node.Identity())
	if err != nil {
		return
	}

	for _, dbContract := range contracts {
		signed, err := objects.Load[*apphost.SignedAppContract](nil, mod.Objects.ReadDefault(), dbContract.ObjectID)
		if err != nil {
			mod.log.Errorv(2, "error loading contract %v: %v", dbContract.ObjectID, err)
			continue
		}
		list = append(list, signed)
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
	signed, err := objects.Load[*apphost.SignedAppContract](ctx, mod.Objects.ReadDefault(), objectID)
	if err != nil {
		return fmt.Errorf("cannot load app contract: %w", err)
	}

	if !mod.isActive(signed) {
		return apphost.ErrInactiveContract
	}

	// save the contract
	err = mod.db.SaveAppContract(&dbAppContract{
		ObjectID:  objectID,
		AppID:     signed.AppID,
		HostID:    signed.HostID,
		StartsAt:  time.Time(signed.StartsAt),
		ExpiresAt: time.Time(signed.ExpiresAt),
	})
	if err != nil {
		return err
	}

	mod.Objects.Receive(&apphost.EventNewAppContract{Contract: signed}, nil)

	return
}

// isActive returns true if the contract is active, i.e., the signatures are valid, and its conditions are met (such
// as start and expiry date).
func (mod *Module) isActive(signed *apphost.SignedAppContract) bool {
	switch {
	case signed.StartsAt.Time().After(time.Now()):
		return false // hasn't started yet

	case signed.ExpiresAt.Time().Before(time.Now()):
		return false // has expired
	case mod.validateSignatures(signed) != nil:
		return false // invalid signatures
	}

	return true
}

func (mod *Module) validateSignatures(signed *apphost.SignedAppContract) error {
	if signed.IsNil() {
		return errors.New("nil contract")
	}
	if signed.HostSig == nil {
		return errors.New("host signature is missing")
	}
	if signed.AppSig == nil {
		return errors.New("app signature is missing")
	}

	hash := signed.SignableHash()
	if hash == nil {
		return errors.New("cannot compute contract hash")
	}

	if err := mod.Crypto.VerifyHashSignature(
		secp256k1.FromIdentity(signed.HostID),
		signed.HostSig,
		hash,
	); err != nil {
		return fmt.Errorf("host signature: %w", err)
	}

	if err := mod.Crypto.VerifyHashSignature(
		secp256k1.FromIdentity(signed.AppID),
		signed.AppSig,
		hash,
	); err != nil {
		return fmt.Errorf("app signature: %w", err)
	}

	return nil
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
		case objectType != apphost.SignedAppContract{}.ObjectType():
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
