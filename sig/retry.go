package sig

import (
	"fmt"
	"math"
	"sync"
	"time"
)

// Retry implements exponential backoff.
//
//	func demo(ctx context.Context) {
//		r, _ := NewRetry(time.Second, time.Minute, 2)
//		for {
//			// try to do something
//			// r.Reset() // reset the retry delay on success
//
//			// on error retry
//			select {
//			case i := <-r.Retry():
//				fmt.Println("retry count", i)
//			case <-ctx.Done():
//				return
//			}
//		}
//	}
type Retry struct {
	mu         sync.Mutex
	retryCount int
	retryChan  chan int
	minDelay   time.Duration
	maxDelay   time.Duration
	nextDelay  time.Duration
	factor     float64
	timer      *time.Timer
	running    bool
}

// NewRetry creates a new Retry instance with provided minimum and maximum delay and growth factor.
func NewRetry(minDelay, maxDelay time.Duration, factor float64) (*Retry, error) {
	switch {
	case minDelay <= 0:
		return nil, fmt.Errorf("invalid argument: minDelay must be greater than 0")
	case maxDelay <= 0:
		return nil, fmt.Errorf("invalid argument: maxDelay must be greater than 0")
	case minDelay > maxDelay:
		return nil, fmt.Errorf("invalid argument: minDelay must be less than or equal to maxDelay")
	case factor < 1.0:
		return nil, fmt.Errorf("invalid argument: factor cannot be less than 1.0")
	}

	r := &Retry{
		nextDelay: minDelay,
		minDelay:  minDelay,
		maxDelay:  maxDelay,
		factor:    factor,
		retryChan: make(chan int, 0),
	}

	return r, nil
}

// Retry returns a channel that receives a value each time a retry should be attempted. The received value
// is the retry counter.
func (r *Retry) Retry() <-chan int {
	r.mu.Lock()
	defer r.mu.Unlock()

	// just return the channel if the clock is already running
	if r.running {
		return r.retryChan
	}
	r.running = true

	// set the timer
	if r.timer == nil {
		r.timer = time.NewTimer(r.nextDelay)
	} else {
		r.timer.Reset(r.nextDelay)
	}

	go func() {
		// trigger the retryCount after the timer fires
		<-r.timer.C

		// lock the retryCount
		r.mu.Lock()
		defer r.mu.Unlock()

		r.trigger()
		r.running = false
	}()

	return r.retryChan
}

// trigger sends a signal to the retry channel and increments the retry counter and delay if successful.
func (r *Retry) trigger() {
	select {
	case r.retryChan <- r.retryCount + 1:
		r.retryCount++
		r.nextDelay = time.Duration(math.Min(float64(r.nextDelay)*r.factor, float64(r.maxDelay)))
	default:
	}
}

// Reset resets the backoff to the initial delay and restarts the sequence.
func (r *Retry) Reset() {
	r.mu.Lock()
	defer r.mu.Unlock()

	// reset the delay
	r.nextDelay = r.minDelay

	// reset the retryCount counter
	r.retryCount = 0

	// reset the timer
	if r.timer != nil {
		r.timer.Stop()
	}
}

// NextDelay returns the next retry delay duration
func (r *Retry) NextDelay() time.Duration {
	r.mu.Lock()
	defer r.mu.Unlock()

	return r.nextDelay
}
