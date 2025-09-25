package udp

import (
	"io"
	"sync"
	"testing"
	"time"
)

func TestRingBufferBlockingWriteRead(t *testing.T) {
	rb := newRingBuffer(10)

	var wg sync.WaitGroup
	wg.Add(2)

	// Writer
	go func() {
		defer wg.Done()
		data := []byte("hello world")
		written, err := rb.WriteAll(data)
		if written != len(data) || err != nil {
			t.Errorf("WriteAll failed: written=%d, err=%v", written, err)
		}
	}()

	// Reader
	go func() {
		defer wg.Done()
		buf := make([]byte, 11)
		read, err := rb.Read(buf)
		if read != 11 || err != nil || string(buf) != "hello world" {
			t.Errorf("Read failed: read=%d, err=%v, buf=%s", read, err, buf)
		}
	}()

	wg.Wait()
}

func TestRingBufferCloseWhileReaderWaiting(t *testing.T) {
	rb := newRingBuffer(10)

	var wg sync.WaitGroup
	wg.Add(1)

	// Reader
	go func() {
		defer wg.Done()
		buf := make([]byte, 10)
		_, err := rb.Read(buf)
		if err != io.EOF {
			t.Errorf("Expected io.EOF, got %v", err)
		}
	}()

	time.Sleep(100 * time.Millisecond) // Ensure reader is waiting
	rb.Close()

	wg.Wait()
}

func TestRingBufferCloseWhileWriterWaiting(t *testing.T) {
	rb := newRingBuffer(5)

	var wg sync.WaitGroup
	wg.Add(1)

	// Writer
	go func() {
		defer wg.Done()
		data := []byte("hello world")
		_, err := rb.WriteAll(data)
		if err != io.ErrClosedPipe {
			t.Errorf("Expected io.ErrClosedPipe, got %v", err)
		}
	}()

	time.Sleep(100 * time.Millisecond) // Ensure writer is waiting
	rb.Close()

	wg.Wait()
}

func TestRingBufferConcurrentProducersConsumers(t *testing.T) {
	rb := newRingBuffer(50)

	var wg sync.WaitGroup
	producers := 5
	consumers := 5
	iterations := 100

	// Producers
	for i := 0; i < producers; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < iterations; j++ {
				rb.WriteAll([]byte{byte(id)})
			}
		}(i)
	}

	// Consumers
	for i := 0; i < consumers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			buf := make([]byte, 1)
			for j := 0; j < iterations; j++ {
				rb.Read(buf)
			}
		}()
	}

	wg.Wait()
}

func TestRingBufferZeroCapacity(t *testing.T) {
	rb := newRingBuffer(0)

	var wg sync.WaitGroup
	wg.Add(1)

	// Writer
	go func() {
		defer wg.Done()
		data := []byte("data")
		_, err := rb.WriteAll(data)
		if err != io.ErrClosedPipe {
			t.Errorf("Expected io.ErrClosedPipe, got %v", err)
		}
	}()

	time.Sleep(100 * time.Millisecond) // Ensure writer is waiting
	rb.Close()

	wg.Wait()
}

func TestRingBufferRaceSafety(t *testing.T) {
	rb := newRingBuffer(100)

	var wg sync.WaitGroup
	producers := 10
	consumers := 10

	// Producers
	for i := 0; i < producers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 1000; j++ {
				rb.TryWrite([]byte{byte(j % 256)})
			}
		}()
	}

	// Consumers
	for i := 0; i < consumers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			buf := make([]byte, 10)
			for j := 0; j < 1000; j++ {
				rb.Read(buf)
			}
		}()
	}

	wg.Wait()
}
