package apphost

import (
	"errors"
	"testing"

	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/log"
	"github.com/cryptopunkscc/astrald/lib/routing"
	"github.com/cryptopunkscc/astrald/mod/user"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

// stubUser satisfies user.Module with a fixed identity; only Identity is exercised.
type stubUser struct {
	id *astral.Identity
}

func (s *stubUser) Ready() <-chan struct{} {
	ch := make(chan struct{})
	close(ch)
	return ch
}
func (s *stubUser) Identity() *astral.Identity                                 { return s.id }
func (s *stubUser) LocalSwarm() []*astral.Identity                             { return nil }
func (s *stubUser) NewMaintainLinkTask(*astral.Identity) user.MaintainLinkTask { return nil }
func (s *stubUser) NewSyncNodesTask(*astral.Identity) user.SyncNodesAction     { return nil }
func (s *stubUser) PushToLocalSwarm(*astral.Context, astral.Object)            {}
func (s *stubUser) Expel(*astral.Context, *astral.Identity) (*user.SignedExpulsion, error) {
	return nil, nil
}

func newTestModule(t *testing.T, u user.Module) *Module {
	t.Helper()

	gdb, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	if err = gdb.AutoMigrate(&dbAccessToken{}); err != nil {
		t.Fatalf("migrate: %v", err)
	}

	mod := &Module{
		db:  &DB{gdb},
		log: log.New(nil),
	}
	mod.User = u
	return mod
}

// routeRequestToken drives OpRequestToken through routing.Op exactly like the
// production router does. enRoute controls whether the query is registered as
// en route; extra is applied to the in-flight query's Extra map.
func routeRequestToken(t *testing.T, mod *Module, enRoute bool, extra map[string]any) error {
	t.Helper()

	op, err := routing.NewOp(mod.OpRequestToken)
	if err != nil {
		t.Fatalf("wrap op: %v", err)
	}

	q := astral.NewQuery(nil, nil, "apphost.request_token")
	inFlight := astral.Launch(q)
	for k, v := range extra {
		inFlight.Extra.Set(k, v)
	}

	if enRoute {
		mod.enRoute.Set(q.Nonce, &queryEnRoute{query: inFlight})
		defer mod.enRoute.Delete(q.Nonce)
	}

	_, err = op.RouteQuery(astral.NewContext(nil), inFlight, &nopWriteCloser{})
	return err
}

func rejectCode(t *testing.T, err error) uint8 {
	t.Helper()

	var rejected *astral.ErrRejected
	if !errors.As(err, &rejected) {
		t.Fatalf("want ErrRejected, got: %v", err)
	}
	return rejected.Code
}

func TestRequestToken_NotEnRoute(t *testing.T) {
	mod := newTestModule(t, &stubUser{id: astral.GenerateIdentity()})

	err := routeRequestToken(t, mod, false, nil)

	if code := rejectCode(t, err); code != 1 {
		t.Fatalf("reject code = %v; want 1", code)
	}
}

func TestRequestToken_NoWebOrigin(t *testing.T) {
	mod := newTestModule(t, &stubUser{id: astral.GenerateIdentity()})

	err := routeRequestToken(t, mod, true, nil)

	if code := rejectCode(t, err); code != 1 {
		t.Fatalf("reject code = %v; want 1", code)
	}
}

func TestRequestToken_UntrustedWebOrigin(t *testing.T) {
	mod := newTestModule(t, &stubUser{id: astral.GenerateIdentity()})

	err := routeRequestToken(t, mod, true, map[string]any{"origin-web": "https://evil.example"})

	if code := rejectCode(t, err); code != 1 {
		t.Fatalf("reject code = %v; want 1", code)
	}
}

func TestRequestToken_NoUserModule(t *testing.T) {
	mod := newTestModule(t, nil)

	err := routeRequestToken(t, mod, true, map[string]any{"origin-web": TrustedWebOrigin})

	if code := rejectCode(t, err); code != 2 {
		t.Fatalf("reject code = %v; want 2", code)
	}
}

func TestRequestToken_NoActiveUser(t *testing.T) {
	mod := newTestModule(t, &stubUser{id: nil})

	err := routeRequestToken(t, mod, true, map[string]any{"origin-web": TrustedWebOrigin})

	if code := rejectCode(t, err); code != 2 {
		t.Fatalf("reject code = %v; want 2", code)
	}
}

func TestRequestToken_TrustedWebOrigin(t *testing.T) {
	userID := astral.GenerateIdentity()
	mod := newTestModule(t, &stubUser{id: userID})

	err := routeRequestToken(t, mod, true, map[string]any{"origin-web": TrustedWebOrigin})

	if err != nil {
		t.Fatalf("query rejected: %v", err)
	}

	tokens, err := mod.ListAccessTokens()
	if err != nil {
		t.Fatalf("list tokens: %v", err)
	}
	if len(tokens) != 1 {
		t.Fatalf("token count = %v; want 1", len(tokens))
	}
	if !tokens[0].Identity.IsEqual(userID) {
		t.Fatalf("token identity = %v; want %v", tokens[0].Identity, userID)
	}
}
