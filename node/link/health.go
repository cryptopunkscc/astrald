package link

import (
	"context"
	"fmt"
	_log "github.com/cryptopunkscc/astrald/log"
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

type Health struct {
	link   *Link
	rtts   []time.Duration
	mu     sync.Mutex
	wakeCh chan struct{}
	log    *_log.Logger
}

func NewHealth(link *Link) *Health {
	return &Health{
		link:   link,
		log:    log.Tag("ping"),
		wakeCh: make(chan struct{}, pingMeasurements),
	}
}

// LastRTT returns link's last measured round trip time
func (health *Health) LastRTT() time.Duration {
	health.mu.Lock()
	defer health.mu.Unlock()

	if len(health.rtts) == 0 {
		return defaultRTT
	}
	return health.rtts[len(health.rtts)-1]
}

func (health *Health) AverageRTT() time.Duration {
	health.mu.Lock()
	defer health.mu.Unlock()

	var count = len(health.rtts)
	if count == 0 {
		return defaultRTT
	}

	var total time.Duration
	for _, d := range health.rtts {
		total += d
	}

	return total / time.Duration(count)
}

func (health *Health) Check() {
	for i := 0; i < pingMeasurements; i++ {
		select {
		case health.wakeCh <- struct{}{}:
		default:
			return
		}
	}
}

func (health *Health) pushRTT(rtt time.Duration) {
	health.mu.Lock()
	defer health.mu.Unlock()

	if health.rtts == nil {
		health.rtts = []time.Duration{rtt}
		return
	}

	health.rtts = append(health.rtts, rtt)

	if len(health.rtts) > pingMeasurements {
		health.rtts = health.rtts[1:]
	}
}

func (health *Health) Run(ctx context.Context) error {
	go func() {
		health.Check()
		for {
			select {
			case <-ctx.Done():
				return
			case <-time.After(pingInterval):
				health.Check()
			}
		}
	}()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()

		case <-health.wakeCh:
		}

		err := health.ping(ctx)
		if err != nil {
			health.link.CloseWithError(fmt.Errorf("health check failed: %w", err))
			return err
		}

		health.log.Infov(2, "ping with %s over %s: last %s avg %s",
			health.link.RemoteIdentity(),
			health.link.Network(),
			health.LastRTT().Round(time.Microsecond),
			health.AverageRTT().Round(time.Microsecond),
		)

		select {
		case <-ctx.Done():
			return ctx.Err()

		case <-time.After(pingCooldown):
		}
	}
}

func (health *Health) ping(ctx context.Context) error {
	resultCh := make(chan error, 1)

	var reader = mux.NewFrameReader()
	localPort, err := health.link.mux.BindAny(reader)
	if err != nil {
		return err
	}

	startTime := time.Now()
	if err := health.link.ctl.WritePing(localPort); err != nil {
		health.link.mux.Unbind(localPort)
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
		health.pushRTT(time.Since(startTime))
		return nil

	case <-time.After(pingTimeout):
		return ErrPingTimeout

	case <-ctx.Done():
		return ctx.Err()
	}
}
