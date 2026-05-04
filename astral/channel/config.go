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
		go func() {
			time.Sleep(t)
			close(ch)
		}()
		config.cancelCh = ch
	}
}
