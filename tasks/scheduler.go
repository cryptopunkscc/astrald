package tasks

import (
	"context"
	"github.com/cryptopunkscc/astrald/log"
	"reflect"
	"sync"
	"sync/atomic"
)

// FIFOScheduler is a cuncurrent task scheduler which uses a basic first-in, first-out queue. Tasks are
// executed as soon as a worker is available, in the order they were added.
type FIFOScheduler struct {
	queue   chan Runner
	workers int
	busy    atomic.Int32
	done    atomic.Int32
	err     atomic.Int32
	running atomic.Bool
}

// Statistics is a struct that contains basic statistics of a Scheduler
type Statistics struct {
	Workers   int // number of workers
	Busy      int // number of busy workers
	DoneCount int // number of tasks that ended successfully
	ErrCount  int // number of tasks that ended with an error
	QueueLen  int // number of tasks in the queue
	QueueCap  int // capacity of the queue
}

// NewFIFOScheduler returns a new instance of FIFOScheduler.
func NewFIFOScheduler(workers int, queueSize int) *FIFOScheduler {
	return &FIFOScheduler{
		queue:   make(chan Runner, queueSize),
		workers: workers,
	}
}

// Run the scheduler for the duration of the context. Returns nil when the context ends.
// Errors: ErrAlreadyRunning
func (s *FIFOScheduler) Run(ctx context.Context) error {
	if !s.running.CompareAndSwap(false, true) {
		return ErrAlreadyRunning
	}
	defer s.running.Store(false)

	var wg sync.WaitGroup

	wg.Add(s.workers)
	for i := 0; i < s.workers; i++ {
		go func() {
			defer wg.Done()
			s.worker(ctx)
		}()
	}
	wg.Wait()

	return nil
}

// Add a task to the scheduler. Retruns ErrQueueOverflow if the queue is full, nil otherwise.
func (s *FIFOScheduler) Add(r Runner) error {
	select {
	case s.queue <- r:
		return nil
	default:
		return ErrQueueOverflow
	}
}

// Stats returns Statistics of the scheduler.
func (s *FIFOScheduler) Stats() Statistics {
	return Statistics{
		Workers:   s.workers,
		Busy:      int(s.busy.Load()),
		DoneCount: int(s.done.Load()),
		ErrCount:  int(s.err.Load()),
		QueueLen:  len(s.queue),
		QueueCap:  cap(s.queue),
	}
}

func (s *FIFOScheduler) worker(ctx context.Context) error {
	for {
		select {
		case <-ctx.Done():
			return nil

		case t := <-s.queue:
			s.busy.Add(1)
			if err := t.Run(ctx); err != nil {
				log.Tag("fifo").Errorv(2, "error running %s: %s", reflect.TypeOf(t), err)
				s.err.Add(1)
			} else {
				s.done.Add(1)
			}
			s.busy.Add(-1)
		}
	}
}
