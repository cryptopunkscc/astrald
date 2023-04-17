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

// Last returns link's last measured round trip time
func (ping *Ping) Last() time.Duration {
	return ping.rtt
}

func (ping *Ping) Run(ctx context.Context) error {
	for {
		rtt, err := ping.ping(ctx)
		if err != nil {
			if err == ErrPingTimeout {
				ping.link.CloseWithError(ErrPingTimeout)
			}
			return err
		}

		ping.rtt = rtt

		select {
		case <-time.After(pingInterval):
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

func (ping *Ping) ping(ctx context.Context) (time.Duration, error) {
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
		return d, nil

	case err := <-errCh:
		return 0, err

	case <-time.After(pingTimeout):
		return 0, ErrPingTimeout

	case <-ctx.Done():
		return 0, ctx.Err()
	}
}
