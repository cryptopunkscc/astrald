package relay

import (
	"context"
	"errors"
	"github.com/cryptopunkscc/astrald/events"
	"github.com/cryptopunkscc/astrald/mod/content"
	"github.com/cryptopunkscc/astrald/mod/objects"
	"github.com/cryptopunkscc/astrald/mod/relay"
)

type IndexerService struct {
	*Module
}

func (srv *IndexerService) Run(ctx context.Context) error {
	go srv.watchNewCerts(ctx)
	go srv.watchPurges(ctx)

	<-ctx.Done()

	return nil
}

func (srv *IndexerService) watchNewCerts(ctx context.Context) {
	for event := range srv.content.Scan(ctx, &content.ScanOpts{Type: relay.CertType}) {
		err := srv.indexData(event.ObjectID)
		switch {
		case err == nil:
			srv.log.Infov(1, "added certificate %v", event.ObjectID)

		case errors.Is(err, relay.ErrCertAlreadyIndexed),
			errors.Is(err, objects.ErrNotFound):
			// nothing

		default:
			srv.log.Errorv(1, "error adding cert %v: %v", event.ObjectID, err)
		}
	}
}

func (srv *IndexerService) watchPurges(ctx context.Context) {
	_ = events.Handle[objects.EventPurged](ctx, srv.node.Events(), func(event objects.EventPurged) error {
		srv.verifyIndex(event.ObjectID)
		return nil
	})
}
