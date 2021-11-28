package sig

import "time"

func After(d time.Duration) Signal {
	sig := New()

	go func() {
		<-time.After(d)
		sig <- struct{}{}
	}()

	return sig
}
