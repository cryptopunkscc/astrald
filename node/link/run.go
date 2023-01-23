package link

import (
	"context"
	"github.com/cryptopunkscc/astrald/log"
	"sync"
)

func (l *Link) Run(ctx context.Context) error {
	ctx, cancel := context.WithCancel(ctx)
	l.Touch() // this resets idle to 0

	var errCh = make(chan error, 3)
	var wg sync.WaitGroup
	wg.Add(3)

	log.Tag("peers").Log("new link with %s over %s",
		l.RemoteIdentity(),
		log.Em(l.Network()),
	)

	l.events.Emit(EventLinkEstablished{Link: l})

	go func() {
		defer wg.Done()
		defer cancel()

		if err := l.processQueries(ctx); err != nil {
			errCh <- err
		}
	}()

	go func() {
		defer wg.Done()
		defer cancel()

		if err := l.monitorPing(ctx); err != nil {
			errCh <- err
		}
	}()

	go func() {
		defer wg.Done()
		defer cancel()

		if err := l.monitorIdle(ctx); err != nil {
			errCh <- err
		}
	}()

	wg.Wait()

	log.Tag("peers").Log("link with %s over %s closed (%s%s%s)",
		l.RemoteIdentity(),
		log.Em(l.Network()),
		log.Red(),
		l.err,
		log.Reset(),
	)

	l.events.Emit(EventLinkClosed{Link: l})

	select {
	case err := <-errCh:
		return err
	default:
		return nil
	}
}
