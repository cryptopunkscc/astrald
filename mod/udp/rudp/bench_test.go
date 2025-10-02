package rudp_test

import (
	"context"
	"net"
	"os"
	"sync/atomic"
	"testing"
	"time"

	"github.com/cryptopunkscc/astrald/astral"
	udpmod "github.com/cryptopunkscc/astrald/mod/udp"
	rudp "github.com/cryptopunkscc/astrald/mod/udp/rudp"
)

var benchDebug = os.Getenv("RUDP_BENCH_DEBUG") != ""
var benchProgress = os.Getenv("RUDP_BENCH_PROGRESS") != ""

// BenchmarkRUDPTransfer10MiB performs a single unidirectional transfer of ~10 MiB
// from client -> server over a single RUDP connection on loopback. Handshake and
// setup are excluded from the timed section. The benchmark runs only one data
// transfer regardless of b.N (subsequent iterations return early) to bound total
// execution time while still reporting throughput via b.SetBytes.
func BenchmarkRUDPTransfer10MiB(b *testing.B) {
	if testing.Short() {
		b.Skip("skip in short mode")
	}

	const totalBytes = 10 * 1024 * 1024 // 10 MiB per iteration
	baseChunkSize := 64 * 1024
	b.SetBytes(totalBytes)
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		// Stop the timer during setup/teardown so only the write phase is measured.
		b.StopTimer()

		b.Logf("[iter %d] setup starting", i)
		baseCtx := astral.NewContext(context.Background())

		cfg := rudp.Config{}
		cfg.Normalize() // simplest model: use normalized defaults

		l, err := rudp.Listen(baseCtx, &net.UDPAddr{IP: net.ParseIP("127.0.0.1"), Port: 0}, cfg, 2*time.Second)
		if err != nil {
			b.Fatalf("listen: %v", err)
		}
		b.Logf("[iter %d] listener=%v", i, l.Addr())

		serverAddr := l.Addr().(*net.UDPAddr)
		ipv4Dest := &net.UDPAddr{IP: net.ParseIP("127.0.0.1"), Port: serverAddr.Port}

		acceptedCh := make(chan *rudp.Conn, 1)
		acceptErrCh := make(chan error, 1)
		go func() {
			acceptCtx, cancel := baseCtx.WithTimeout(3 * time.Second)
			defer cancel()
			c, aerr := l.Accept(acceptCtx)
			if aerr != nil {
				acceptErrCh <- aerr
				return
			}
			acceptedCh <- c
		}()

		udpConn, err := net.DialUDP("udp4", nil, ipv4Dest)
		if err != nil {
			udpConn, err = net.DialUDP("udp", nil, ipv4Dest)
		}
		if err != nil {
			l.Close()
			b.Fatalf("dial: %v", err)
		}
		b.Logf("[iter %d] client local=%v remote=%v", i, udpConn.LocalAddr(), udpConn.RemoteAddr())

		localEP, _ := udpmod.ParseEndpoint(udpConn.LocalAddr().String())
		remoteEP, _ := udpmod.ParseEndpoint(udpConn.RemoteAddr().String())
		outConn, err := rudp.NewConn(udpConn, localEP, remoteEP, cfg, true, nil, baseCtx)
		if err != nil {
			udpConn.Close()
			l.Close()
			b.Fatalf("newconn: %v", err)
		}

		hCtx, hCancel := baseCtx.WithTimeout(2 * time.Second)
		if err := outConn.StartClientHandshake(hCtx); err != nil {
			hCancel()
			outConn.Close()
			udpConn.Close()
			l.Close()
			b.Fatalf("handshake: %v", err)
		}
		hCancel()
		b.Logf("[iter %d] handshake complete local=%v remote=%v", i, udpConn.LocalAddr(), udpConn.RemoteAddr())

		var serverConn *rudp.Conn
		select {
		case serverConn = <-acceptedCh:
		case aerr := <-acceptErrCh:
			outConn.Close()
			udpConn.Close()
			l.Close()
			b.Fatalf("accept: %v", aerr)
		case <-time.After(3 * time.Second):
			outConn.Close()
			udpConn.Close()
			l.Close()
			b.Fatalf("accept timeout")
		}
		if serverConn == nil {
			outConn.Close()
			udpConn.Close()
			l.Close()
			b.Fatalf("nil server conn")
		}
		b.Logf("[iter %d] server accepted remote=%v", i, serverConn.RemoteEndpoint())

		// Determine a safe chunk size that will not block on first write:
		// send ring buffer size = MaxWindowBytes * 2. Choose <= MaxWindowBytes /2 for margin.
		safeChunk := cfg.MaxWindowPackets * cfg.MaxSegmentSize / 2
		if safeChunk < cfg.MaxSegmentSize {
			safeChunk = cfg.MaxSegmentSize
		}
		if baseChunkSize > safeChunk {
			if benchDebug {
				b.Logf("[iter %d] adjusting chunk size from %d to %d ("+
					"MaxWindowBytes=%d)", i, baseChunkSize, safeChunk, cfg.MaxWindowPackets)
			}
		}
		chunkSize := safeChunk

		// Prepare transfer
		chunk := make([]byte, chunkSize)
		for j := range chunk {
			chunk[j] = byte(j)
		}
		var received int64
		readDone := make(chan struct{})
		var readErr atomic.Value
		go func() {
			buf := make([]byte, 128*1024)
			for atomic.LoadInt64(&received) < int64(totalBytes) {
				n, err := serverConn.Read(buf)
				if err != nil {
					readErr.Store(err)
					break
				}
				if n > 0 {
					atomic.AddInt64(&received, int64(n))
				}
			}
			close(readDone)
		}()

		b.Logf("[iter %d] starting timed transfer", i)
		writeStart := time.Now()
		// Start the timer only for the write phase (do not reset cumulative timer)
		b.StartTimer()
		sent := 0
		nextMilestone := 1 * 1024 * 1024
		for sent < totalBytes {
			toWrite := chunkSize
			if rem := totalBytes - sent; rem < toWrite {
				toWrite = rem
			}
			if _, err := outConn.Write(chunk[:toWrite]); err != nil {
				b.Fatalf("write err after %d bytes: %v", sent, err)
			}
			sent += toWrite
			if benchProgress && sent >= nextMilestone {
				b.Logf("[iter %d] progress sent=%d / %d (%.2f%%)", i, sent, totalBytes, 100*float64(sent)/float64(totalBytes))
				nextMilestone += 1 * 1024 * 1024
			}
		}
		b.StopTimer()
		elapsedWrite := time.Since(writeStart)

		select {
		case <-readDone:
		case <-time.After(5 * time.Second):
			b.Fatalf("read timeout received=%d", atomic.LoadInt64(&received))
		}
		if v := readErr.Load(); v != nil {
			b.Fatalf("server read error: %v", v.(error))
		}
		if got := atomic.LoadInt64(&received); got != int64(totalBytes) {
			b.Fatalf("recv mismatch got=%d want=%d", got, totalBytes)
		}

		mbps := (float64(totalBytes) / (1024 * 1024)) / elapsedWrite.Seconds()
		b.Logf("[iter %d] transfer complete bytes=%d duration=%v throughput=%.2f MiB/s", i, totalBytes, elapsedWrite, mbps)

		serverConn.Close()
		outConn.Close()
		udpConn.Close()
		l.Close()
		b.Logf("[iter %d] cleanup done", i)
	}
}
