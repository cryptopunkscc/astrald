package link

import (
	"context"
	"fmt"
	"github.com/cryptopunkscc/astrald/mux"
	"io"
	"sync"
	"time"
)

const pingInterval = 30 * time.Minute
const pingCooldown = time.Second
const pingTimeout = 10 * time.Second
const pingMeasurements = 3
const defaultRTT = 999 * time.Second

type Ping struct {
	link   *Link
	rtts   []time.Duration
	mu     sync.Mutex
	wakeCh chan struct{}
}

func NewPing(link *Link) *Ping {
	return &Ping{
		link:   link,
		wakeCh: make(chan struct{}, pingMeasurements),
	}
}

// Last returns link's last measured round trip time
func (ping *Ping) Last() time.Duration {
	ping.mu.Lock()
	defer ping.mu.Unlock()

	if len(ping.rtts) == 0 {
		return defaultRTT
	}
	return ping.rtts[len(ping.rtts)-1]
}

func (ping *Ping) Average() time.Duration {
	ping.mu.Lock()
	defer ping.mu.Unlock()

	var count = len(ping.rtts)
	if count == 0 {
		return defaultRTT
	}

	var total time.Duration
	for _, d := range ping.rtts {
		total += d
	}

	return total / time.Duration(count)
}

func (ping *Ping) Check() {
	for i := 0; i < pingMeasurements; i++ {
		select {
		case ping.wakeCh <- struct{}{}:
		default:
			return
		}
	}
}

func (ping *Ping) pushRTT(rtt time.Duration) {
	ping.mu.Lock()
	defer ping.mu.Unlock()

	if ping.rtts == nil {
		ping.rtts = []time.Duration{rtt}
		return
	}

	ping.rtts = append(ping.rtts, rtt)

	if len(ping.rtts) > pingMeasurements {
		ping.rtts = ping.rtts[1:]
	}
}

func (ping *Ping) Run(ctx context.Context) error {
	go func() {
		ping.Check()
		for {
			select {
			case <-ctx.Done():
				return
			case <-time.After(pingInterval):
				ping.Check()
			}
		}
	}()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()

		case <-ping.wakeCh:
		}

		err := ping.ping(ctx)
		if err != nil {
			ping.link.CloseWithError(err)
			return err
		}

		log.Infov(2, "ping with %s: last %s avg %s",
			ping.link.RemoteIdentity(),
			ping.Last().Round(time.Microsecond),
			ping.Average().Round(time.Microsecond),
		)

		select {
		case <-ctx.Done():
			return ctx.Err()

		case <-time.After(pingCooldown):
		}
	}
}

func (ping *Ping) ping(ctx context.Context) error {
	resultCh := make(chan error, 1)

	var reader = mux.NewFrameReader()
	localPort, err := ping.link.mux.BindAny(reader)
	if err != nil {
		return err
	}

	startTime := time.Now()
	if err := ping.link.ctl.WritePing(localPort); err != nil {
		ping.link.mux.Unbind(localPort)
		return err
	}

	go func() {
		var buf [1]byte
		_, err := reader.Read(buf[:])
		if err != io.EOF {
			resultCh <- fmt.Errorf("unexpected ping response: %w", err)
			return
		}
		resultCh <- nil
	}()

	select {
	case err := <-resultCh:
		if err != nil {
			return err
		}
		ping.pushRTT(time.Since(startTime))
		return nil

	case <-time.After(pingTimeout):
		return ErrPingTimeout

	case <-ctx.Done():
		return ctx.Err()
	}
}
