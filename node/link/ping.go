package link

import (
	"context"
	"time"
)

type Ping struct {
	link *Link
	rtt  time.Duration
}

func NewPing(link *Link) *Ping {
	return &Ping{
		link: link,
		rtt:  999 * time.Second, // assume super slow before first ping
	}
}

// RTT returns link's round trip time
func (ping *Ping) RTT() time.Duration {
	return ping.rtt
}

func (ping *Ping) Run(ctx context.Context) error {
	for {
		if err := ping.ping(ctx); err != nil {
			return err
		}

		select {
		case <-time.After(pingInterval):
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

func (ping *Ping) ping(ctx context.Context) error {
	pingCh := make(chan time.Duration, 1)
	errCh := make(chan error, 1)

	go func() {
		startTime := time.Now()
		conn, err := ping.link.Query(ctx, ".ping")
		roundTrip := time.Since(startTime)

		switch err {
		case nil:
			conn.Close()
			fallthrough

		case ErrRejected:
			pingCh <- roundTrip

		default:
			errCh <- err
		}
	}()

	select {
	case d := <-pingCh:
		ping.rtt = d

	case err := <-errCh:
		return err

	case <-time.After(pingTimeout):
		return ErrPingTimeout

	case <-ctx.Done():
		return ctx.Err()
	}

	return nil
}
