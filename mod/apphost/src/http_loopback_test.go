package apphost

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

// TestHTTPServer_LoopbackOnly verifies the HTTP query/object surface is refused
// for non-loopback callers and admitted (past the loopback gate, to token auth)
// for loopback callers. The WebSocket path is guarded separately in ws_server.go.
func TestHTTPServer_LoopbackOnly(t *testing.T) {
	mod := newTestModule(t, nil)
	srv := &HTTPServer{Module: mod}

	cases := []struct {
		name       string
		remoteAddr string
		wantStatus int
	}{
		{"LAN address refused", "192.168.1.50:40000", http.StatusForbidden},
		{"public address refused", "203.0.113.7:40000", http.StatusForbidden},
		// loopback passes the gate, then fails token auth (no Authorization header)
		{"loopback reaches auth", "127.0.0.1:40000", http.StatusUnauthorized},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/user.info", nil)
			req.RemoteAddr = c.remoteAddr
			rec := httptest.NewRecorder()

			srv.ServeHTTP(rec, req)

			if rec.Code != c.wantStatus {
				t.Fatalf("status = %d; want %d", rec.Code, c.wantStatus)
			}
		})
	}
}
