package apphost

import (
	"context"
	"crypto/rand"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/coder/websocket"
)

// roundTripWSConn opens a WS pair using msgType frames and verifies that bytes written on
// one side appear verbatim on the other.
func roundTripWSConn(t *testing.T, msgType websocket.MessageType) {
	t.Helper()

	payload := make([]byte, 64*1024)
	if _, err := rand.Read(payload); err != nil {
		t.Fatalf("rand: %v", err)
	}

	serverDone := make(chan error, 1)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		c, err := websocket.Accept(w, r, &websocket.AcceptOptions{
			InsecureSkipVerify: true,
		})
		if err != nil {
			serverDone <- err
			return
		}

		conn := newWSConn(r.Context(), c, msgType)
		defer conn.Close()

		// Echo: read everything, write it back.
		buf := make([]byte, len(payload))
		if _, err := io.ReadFull(conn, buf); err != nil {
			serverDone <- err
			return
		}
		if _, err := conn.Write(buf); err != nil {
			serverDone <- err
			return
		}

		serverDone <- nil
	}))
	defer srv.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	wsURL := strings.Replace(srv.URL, "http://", "ws://", 1)
	c, _, err := websocket.Dial(ctx, wsURL, nil)
	if err != nil {
		t.Fatalf("dial: %v", err)
	}
	defer c.CloseNow()

	clientConn := websocket.NetConn(ctx, c, msgType)

	if _, err := clientConn.Write(payload); err != nil {
		t.Fatalf("client write: %v", err)
	}

	echo := make([]byte, len(payload))
	if _, err := io.ReadFull(clientConn, echo); err != nil {
		t.Fatalf("client read: %v", err)
	}

	for i := range payload {
		if echo[i] != payload[i] {
			t.Fatalf("byte %d differs: got %02x, want %02x", i, echo[i], payload[i])
		}
	}

	clientConn.Close()
	if err := <-serverDone; err != nil {
		t.Fatalf("server: %v", err)
	}
}

func TestWSConn_BinaryRoundTrip(t *testing.T) {
	roundTripWSConn(t, websocket.MessageBinary)
}

func TestWSConn_TextRoundTrip(t *testing.T) {
	// Text frames carry arbitrary bytes for our purposes too — coder/websocket doesn't
	// enforce UTF-8 on the read side. We use small ASCII bytes to keep it strict.
	roundTripWSConn(t, websocket.MessageText)
}

func TestWSConn_CloseIdempotent(t *testing.T) {
	type result struct {
		firstErr, secondErr error
		acceptErr           error
	}
	resultCh := make(chan result, 1)
	ready := make(chan struct{})

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		c, err := websocket.Accept(w, r, &websocket.AcceptOptions{InsecureSkipVerify: true})
		if err != nil {
			resultCh <- result{acceptErr: err}
			close(ready)
			return
		}
		conn := newWSConn(r.Context(), c, websocket.MessageBinary)
		close(ready)
		// Wait until the client side has initiated close so our Close() doesn't
		// block on the close handshake.
		_, _, _ = c.Reader(r.Context())
		first := conn.Close()
		second := conn.Close()
		resultCh <- result{firstErr: first, secondErr: second}
	}))
	defer srv.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	wsURL := strings.Replace(srv.URL, "http://", "ws://", 1)
	c, _, err := websocket.Dial(ctx, wsURL, nil)
	if err != nil {
		t.Fatalf("dial: %v", err)
	}

	<-ready
	c.Close(websocket.StatusNormalClosure, "test done")

	select {
	case r := <-resultCh:
		if r.acceptErr != nil {
			t.Fatalf("accept: %v", r.acceptErr)
		}
		// Idempotency: the second Close must return the same cached result as the first.
		// (Either may return an error if the peer closed first — what matters is that
		// the second call doesn't attempt to close the underlying conn again, which
		// would surface as a different error path.)
		firstStr := errString(r.firstErr)
		secondStr := errString(r.secondErr)
		if firstStr != secondStr {
			t.Errorf("close idempotency broken: first=%q second=%q", firstStr, secondStr)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("handler did not complete")
	}
}

func errString(err error) string {
	if err == nil {
		return ""
	}
	return err.Error()
}
