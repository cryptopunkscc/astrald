package channel

import (
	"context"
	"time"
)

type ConfigFunc func(*Config)

type Config struct {
	fmtIn, fmtOut string
	allowUnparsed bool
	lockWrites    bool
	cancelCh      <-chan struct{}
	// cancelStop releases resources backing cancelCh (e.g. WithTimeout's
	// timer). Consumers of cancelCh call it once when they're done reading.
	cancelStop func()
}

func WithInputFormat(fmt string) func(*Config) {
	return func(config *Config) {
		config.fmtIn = fmt
	}
}

func WithOutputFormat(fmt string) func(*Config) {
	return func(config *Config) {
		config.fmtOut = fmt
	}
}

func WithFormats(fmtIn, fmtOut string) func(*Config) {
	return func(config *Config) {
		config.fmtIn = fmtIn
		config.fmtOut = fmtOut
	}
}

func WithFormat(fmt string) func(*Config) {
	return WithFormats(fmt, fmt)
}

func AllowUnparsed(b bool) func(*Config) {
	return func(config *Config) {
		config.allowUnparsed = b
	}
}

func WithContext(ctx context.Context) func(*Config) {
	return func(config *Config) {
		config.cancelCh = ctx.Done()
	}
}

// WithLockedWrites puts the calls to concrete Senders in a mutex. Use only
// with New() (individual senders unsupported).
func WithLockedWrites() func(*Config) {
	return func(config *Config) {
		config.lockWrites = true
	}
}

func WithTimeout(t time.Duration) func(*Config) {
	return func(config *Config) {
		ch := make(chan struct{})
		timer := time.AfterFunc(t, func() { close(ch) })
		config.cancelCh = ch
		config.cancelStop = func() { timer.Stop() }
	}
}
