package gateway

import (
	"io"
	"time"
)

func pipe(a, b io.ReadWriteCloser) {
	const idle = 30 * time.Second

	done := make(chan struct{}, 2)

	forward := func(dst, src io.ReadWriteCloser) {
		// note: sync.Pool could reduce per-connection allocations under high concurrency (pattern used by nginx, envoy, traefik)
		buf := make([]byte, 32*1024)
		srcD, srcOk := src.(deadliner)
		dstD, dstOk := dst.(deadliner)
		for {
			if srcOk {
				srcD.SetReadDeadline(time.Now().Add(idle))
			}
			n, err := src.Read(buf)
			if n > 0 {
				if dstOk {
					dstD.SetWriteDeadline(time.Now().Add(idle))
				}
				if _, werr := dst.Write(buf[:n]); werr != nil {
					break
				}
			}
			if err != nil {
				break
			}
		}
		done <- struct{}{}
	}

	go forward(a, b)
	go forward(b, a)

	<-done
	a.Close()
	b.Close()
}
