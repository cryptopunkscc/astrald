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
	// text mode flushes only complete lines; make the payload one giant line
	for i, b := range payload[:len(payload)-1] {
		if b == '\n' {
			payload[i] = ' '
		}
	}
	payload[len(payload)-1] = '\n'

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
	// Text mode re-frames on newlines, so only newline-terminated data is
	// flushed; roundTripWSConn ends its payload with '\n'.
	roundTripWSConn(t, websocket.MessageText)
}

// TestWSConn_TextLineFraming verifies the astral.json.v1 invariant: one
// newline-terminated line per text frame, regardless of how writes chunk the
// byte stream.
func TestWSConn_TextLineFraming(t *testing.T) {
	serverDone := make(chan error, 1)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		c, err := websocket.Accept(w, r, &websocket.AcceptOptions{InsecureSkipVerify: true})
		if err != nil {
			serverDone <- err
			return
		}

		conn := newWSConn(r.Context(), c, websocket.MessageText)
		defer conn.Close()

		// two complete lines in a single Write -> two frames
		if _, err := conn.Write([]byte("{\"a\":1}\n{\"b\":2}\n")); err != nil {
			serverDone <- err
			return
		}
		// one line spread over three Writes -> one frame
		for _, part := range []string{"{\"c\"", ":3}", "\n"} {
			if _, err := conn.Write([]byte(part)); err != nil {
				serverDone <- err
				return
			}
		}
		serverDone <- nil

		// hold the conn open until the client is done reading
		_, _, _ = c.Reader(r.Context())
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

	want := []string{"{\"a\":1}\n", "{\"b\":2}\n", "{\"c\":3}\n"}
	for i, expected := range want {
		typ, data, err := c.Read(ctx)
		if err != nil {
			t.Fatalf("frame %d: read: %v", i, err)
		}
		if typ != websocket.MessageText {
			t.Fatalf("frame %d: type = %v, want text", i, typ)
		}
		if string(data) != expected {
			t.Fatalf("frame %d = %q, want %q", i, data, expected)
		}
	}

	if err := <-serverDone; err != nil {
		t.Fatalf("server: %v", err)
	}
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
