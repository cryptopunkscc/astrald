package objects

import (
	"errors"
	"time"

	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/log"
	"github.com/cryptopunkscc/astrald/lib/astrald"
	"github.com/cryptopunkscc/astrald/mod/objects"
	objectscli "github.com/cryptopunkscc/astrald/mod/objects/client"
	"github.com/cryptopunkscc/astrald/sig"
)

type AppDescriber struct {
	mod     *Module
	id      *astral.Identity
	client  *objectscli.Client
	log     *log.Logger
	timeout time.Duration
}

func NewAppDescriber(mod *Module, id *astral.Identity) *AppDescriber {
	return &AppDescriber{
		mod:     mod,
		id:      id,
		client:  objectscli.New(id, astrald.Default()),
		log:     mod.log.AppendTag(log.Tag(id.Fingerprint())),
		timeout: defaultAppDiscovererTimeout,
	}
}

func (d *AppDescriber) SourceIdentity() *astral.Identity { return d.id }

func (d *AppDescriber) DescribeObject(ctx *astral.Context, id *astral.ObjectID) (<-chan *objects.Descriptor, error) {
	ctx, cancel := ctx.WithTimeout(d.timeout)

	in, errPtr := d.client.Describe(ctx, id)
	if in == nil {
		cancel()
		if errPtr != nil && *errPtr != nil {
			return nil, *errPtr
		}
		return nil, errors.New("app describe returned no stream")
	}

	out := make(chan *objects.Descriptor)
	go func() {
		defer cancel()
		defer close(out)

		for {
			descriptor, ok, err := sig.RecvOk(ctx, in)
			if err != nil {
				d.log.Errorv(1, "app describer: %v", err)
				return
			}
			if !ok {
				break
			}
			if descriptor == nil {
				continue
			}

			descriptor.SourceID = d.id
			if err := sig.Send(ctx, out, descriptor); err != nil {
				return
			}
		}

		if errPtr != nil && *errPtr != nil {
			d.log.Errorv(1, "app describer: %v", *errPtr)
		}
	}()

	return out, nil
}
