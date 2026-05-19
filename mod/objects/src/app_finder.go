package objects

import (
	"errors"
	"time"

	"github.com/cryptopunkscc/astrald/astral"
	log "github.com/cryptopunkscc/astrald/astral/log"
	"github.com/cryptopunkscc/astrald/lib/astrald"
	objectscli "github.com/cryptopunkscc/astrald/mod/objects/client"
	"github.com/cryptopunkscc/astrald/sig"
)

type AppFinder struct {
	mod     *Module
	id      *astral.Identity
	client  *objectscli.Client
	log     *log.Logger
	timeout time.Duration
}

func NewAppFinder(mod *Module, id *astral.Identity) *AppFinder {
	return &AppFinder{
		mod:     mod,
		id:      id,
		client:  objectscli.New(id, astrald.Default()),
		log:     mod.log.AppendTag(log.Tag(id.Fingerprint())),
		timeout: defaultAppDiscovererTimeout,
	}
}

func (f *AppFinder) SourceIdentity() *astral.Identity { return f.id }

func (f *AppFinder) FindObject(ctx *astral.Context, id *astral.ObjectID) (<-chan *astral.Identity, error) {
	providerCtx, cancel := ctx.WithTimeout(f.timeout)

	providers, errPtr := f.client.Find(providerCtx, id)
	if providers == nil {
		cancel()
		if errPtr != nil && *errPtr != nil {
			return nil, *errPtr
		}
		return nil, errors.New("app find returned no stream")
	}

	out := make(chan *astral.Identity)
	go func() {
		defer cancel()
		defer close(out)

		for {
			provider, ok, err := sig.RecvOk(providerCtx, providers)
			if err != nil {
				f.log.Errorv(1, "app finder: %v", err)
				return
			}
			if !ok {
				break
			}

			if err := sig.Send(providerCtx, out, provider); err != nil {
				return
			}
		}

		if errPtr != nil && *errPtr != nil {
			f.log.Errorv(1, "app finder: %v", *errPtr)
		}
	}()

	return out, nil
}
