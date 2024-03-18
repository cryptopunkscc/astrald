package srv

import (
	"context"
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/data"
	"github.com/cryptopunkscc/astrald/lib/astral"
	client "github.com/cryptopunkscc/astrald/lib/storage"
	"github.com/cryptopunkscc/astrald/mod/storage"
	proto "github.com/cryptopunkscc/astrald/mod/storage/srv"
	"github.com/go-playground/assert/v2"
	"log"
	"path"
	"strings"
	"testing"
	"time"
)

const testStoragePort = proto.Port + "test."

func TestStorageService(t *testing.T) {
	m := newTestStorageModule()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	service := NewService(m, nil)
	service.port = testStoragePort
	service.register = service.registerTestRoute
	if err := service.Run(ctx); err != nil {
		t.Fatal(err)
	}

	c := client.NewClient(id.Anyone).Port(testStoragePort)

	t.Run("ReadAll", func(t *testing.T) {
		dataID := data.ID{Size: 1}
		opts := &storage.OpenOpts{Offset: 1}
		bytes, err := c.ReadAll(dataID, opts)
		m.verifyEq(t, err, dataID, opts, bytes)
	})

	t.Run("ReadAll#Error", func(t *testing.T) {
		dataID := data.ID{Size: 1}
		opts := &storage.OpenOpts{Virtual: true}
		bytes, err := c.ReadAll(dataID, opts)
		assert.Equal(t, nil, bytes)
		m.verifyEq(t, nil, err)
	})

	t.Run("Put", func(t *testing.T) {
		bytes := []byte{1}
		opts := &storage.CreateOpts{Alloc: 1}
		put, err := c.Put(bytes, opts)
		m.verifyEq(t, err, bytes, opts, put)
	})

	t.Run("Purge", func(t *testing.T) {
		dataID := data.ID{Size: 1}
		var opts *storage.PurgeOpts
		l, err := c.Purge(dataID, opts)
		m.verifyEq(t, err, dataID, opts, l)
	})

	t.Run("Create", func(t *testing.T) {
		opts := &storage.CreateOpts{Alloc: 1}
		writer, err := c.Create(opts)
		m.verifyEq(t, err, opts)

		t.Run("Write", func(t *testing.T) {
			b := make([]byte, 16)
			l, err := writer.Write(b)
			m.verifyEq(t, err, b, l)
		})

		t.Run("Commit", func(t *testing.T) {
			d, err := writer.Commit()
			m.verifyEq(t, err, d)
		})
	})

	t.Run("Create", func(t *testing.T) {
		opts := &storage.CreateOpts{Alloc: 1}
		writer, err := c.Create(opts)
		m.verifyEq(t, err, opts)

		t.Run("Discard", func(tt *testing.T) {
			err := writer.Discard()
			m.verifyEq(tt, err)
		})
	})

	t.Run("Open", func(t *testing.T) {
		var f id.Filter
		dataID := data.ID{Size: 1}
		opts := &storage.OpenOpts{Offset: 1, IdentityFilter: f}
		reader, err := c.Open(dataID, opts)
		m.verifyEq(t, err, dataID, opts)
		t.Cleanup(func() { _ = reader.Close() })

		t.Run("Info", func(t *testing.T) {
			info := reader.Info()
			m.verifyEq(t, err, info)
		})

		t.Run("Seek", func(t *testing.T) {
			offset := int64(1)
			whence := 1
			seek, err := reader.Seek(offset, whence)
			m.verifyEq(t, err, offset, whence, seek)
		})

		t.Run("Read", func(t *testing.T) {
			b := make([]byte, 16)
			read, err := reader.Read(b)
			m.verifyEq(t, err, b, read)
		})
	})

	t.Run("Open", func(t *testing.T) {
		var f id.Filter = func(identity id.Identity) bool {
			m.notify("idFilter", identity, true)
			return true
		}
		dataID := data.ID{Size: 1}
		opts := &storage.OpenOpts{Offset: 1, IdentityFilter: f}
		reader, err := c.Open(dataID, opts)

		m.verify(t, err, func(left []any, right []any) {
			left[2] = right[2]
			assert.Equal(t, left, right)
			if f != nil {
				f = right[2].(*storage.OpenOpts).IdentityFilter
			}
		}, dataID, opts)
		t.Cleanup(func() { _ = reader.Close() })

		t.Run("idFilter", func(t *testing.T) {
			i, err := id.GenerateIdentity()
			if err != nil {
				t.Fatal(err)
			}
			i, err = id.ParsePublicKeyHex(i.PublicKeyHex())
			if err != nil {
				t.Fatal(err)
			}
			b := f(i)
			m.verifyEq(t, nil, i, b)
		})
	})
}

func (s *Service) registerTestRoute(ctx context.Context, route string) (err error) {
	l, err := astral.Register(route)
	if err != nil {
		return
	}
	go func() {
		<-ctx.Done()
		_ = l.Close()
	}()
	go func() {
		for qd := range l.QueryCh() {
			go func(qd *astral.QueryData) {
				cmd := s.parseCmd(qd.Query())
				h, ok := s.handlers[cmd]

				if !ok {
					_ = qd.Reject()
					return
				}

				conn, err := qd.Accept()
				if err != nil {
					return
				}
				defer conn.Close()

				ctx := Context{
					Module:   s.Module,
					Conn:     conn,
					RemoteID: qd.RemoteIdentity(),
				}

				if err := h.handle(ctx, conn, s.unmarshal, s.encoder(conn), qd.Query()); err != nil {
					log.Println(err)
				}
			}(qd)
		}
	}()
	return
}

type testStorageModule struct {
	events chan []any
}

func newTestStorageModule() *testStorageModule {
	return &testStorageModule{
		events: make(chan []any, 1024),
	}
}

func (t *testStorageModule) Open(dataID data.ID, opts *storage.OpenOpts) (storage.Reader, error) {
	t.notify("Open", dataID, opts)
	return &testStorageReader{t}, nil
}

type testStorageReader struct {
	*testStorageModule
}

func (t *testStorageReader) Read(p []byte) (n int, err error) {
	n = cap(p)
	t.notify("Read", p, n)
	return
}

func (t *testStorageReader) Seek(offset int64, whence int) (i int64, err error) {
	i = 1
	t.notify("Seek", offset, whence, i)
	return
}

func (t *testStorageReader) Close() error {
	t.notify("Close")
	return nil
}

func (t *testStorageReader) Info() (info *storage.ReaderInfo) {
	info = &storage.ReaderInfo{Name: "Info"}
	t.notify("Info", info)
	return
}

func (t *testStorageModule) Create(opts *storage.CreateOpts) (storage.Writer, error) {
	t.notify("Create", opts)
	return &storageWriter{t}, nil
}

type storageWriter struct {
	*testStorageModule
}

func (s *storageWriter) Write(p []byte) (n int, err error) {
	n = cap(p)
	s.notify("Write", p, n)
	return
}

func (s *storageWriter) Commit() (d data.ID, err error) {
	d = data.ID{Size: 1}
	d.Hash[0] = 1
	s.notify("Commit", d)
	return
}

func (s *storageWriter) Discard() error {
	s.notify("Discard")
	return nil
}

func (t *testStorageModule) Purge(dataID data.ID, opts *storage.PurgeOpts) (l int, err error) {
	l = 1
	t.notify("Purge", dataID, opts, l)
	return
}

func (t *testStorageModule) ReadAll(id data.ID, opts *storage.OpenOpts) (b []byte, err error) {
	b = []byte{1}
	if opts.Virtual {
		err = &proto.ReadAllResp{Response: proto.Response{Err: "test error"}}
		t.notify("ReadAll", err)
	} else {
		t.notify("ReadAll", id, opts, b)
	}
	return
}

func (t *testStorageModule) Put(bytes []byte, opts *storage.CreateOpts) (d data.ID, err error) {
	d = data.ID{Size: 1}
	t.notify("Put", bytes, opts, d)
	return
}

func (t *testStorageModule) AddOpener(name string, opener storage.Opener, priority int) error {
	t.notify("AddOpener", name, opener, priority)
	return nil
}

func (t *testStorageModule) RemoveOpener(name string) error {
	t.notify("RemoveOpener", name)
	return nil
}

func (t *testStorageModule) AddCreator(name string, creator storage.Creator, priority int) error {
	t.notify("AddCreator", name, creator, priority)
	return nil
}

func (t *testStorageModule) RemoveCreator(name string) error {
	t.notify("RemoveCreator", name)
	return nil
}

func (t *testStorageModule) AddPurger(name string, purger storage.Purger) error {
	t.notify("AddPurger", name, purger)
	return nil
}

func (t *testStorageModule) RemovePurger(name string) error {
	t.notify("RemovePurger", name)
	return nil
}

func (t *testStorageModule) notify(name string, args ...any) {
	arr := append([]any{name}, args...)
	t.events <- arr
}

func (t *testStorageModule) verifyEq(tt *testing.T, err error, args ...any) {
	t.verify(tt, err, func(left []any, right []any) { assert.EqualSkip(tt, 4, left, right) }, args...)
}

func (t *testStorageModule) verify(
	tt *testing.T,
	err error,
	check func(left []any, right []any),
	args ...any,
) {
	tt.Helper()
	name := strings.Split(path.Base(tt.Name()), "#")[0]
	if err != nil {
		tt.Error(args, err)
	}
	left := append([]any{name}, args...)
	var right []any
	tick := time.Tick(3 * time.Second)
await:
	for {
		select {
		case <-tick:
			tt.Errorf("verify timeout for %v", left)
			return
		case right = <-t.events:
			if right[0] == name {
				break await
			}
			t.events <- right
			time.Sleep(5 * time.Millisecond)
		}
	}
	check(left, right)
	return
}
