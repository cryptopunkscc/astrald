package link

import (
	"context"
	"github.com/cryptopunkscc/astrald/link"
	"time"
)

// Ping returns link's round trip time
func (l *Link) Ping() time.Duration {
	return l.roundtrip
}

func (l *Link) monitorPing(ctx context.Context) error {
	for {
		if err := l.ping(ctx); err != nil {
			l.Close()
			return err
		}

		// if the connection is active, we want to monitor ping more often
		interval := idlePingInterval
		if l.Activity.Idle() == 0 {
			interval = activePingInterval
		}

		// wait between pings
		select {
		case <-time.After(interval):

		case <-ctx.Done():
			return ctx.Err()

		case <-l.Wait():
			return nil
		}
	}
}

func (l *Link) ping(ctx context.Context) error {
	pingCh := make(chan time.Duration, 1)
	errCh := make(chan error, 1)

	go func() {
		startTime := time.Now()
		conn, err := l.Query(ctx, ".ping")
		roundTrip := time.Since(startTime)

		switch err {
		case nil:
			conn.Close()
			fallthrough
		case link.ErrRejected:
			pingCh <- roundTrip
		default:
			errCh <- err
		}
	}()

	select {
	case d := <-pingCh:
		l.roundtrip = d

	case err := <-errCh:
		return err

	case <-time.After(pingTimeout):
		l.err = ErrPingTimeout
		return ErrPingTimeout

	case <-ctx.Done():
		return ctx.Err()
	}

	return nil
}
