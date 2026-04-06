package apphost

import (
	"errors"
	"fmt"
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
