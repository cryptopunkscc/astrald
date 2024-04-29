package srv

import (
	"context"
	"errors"
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/data"
	client "github.com/cryptopunkscc/astrald/lib/storage"
	"github.com/cryptopunkscc/astrald/mod/storage"
	proto "github.com/cryptopunkscc/astrald/mod/storage/srv"
	jrpc "github.com/cryptopunkscc/go-apphost-jrpc"
	"github.com/go-playground/assert/v2"
	"log"
	"path"
	"strings"
	"testing"
	"time"
)

const testStoragePort = proto.Port + ".test"

func TestStorageService(t *testing.T) {
	m := newTestStorageModule()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	app := jrpc.NewApp(testStoragePort)
	app.Logger(log.New(log.Writer(), "service ", 0))
	app.Interface(NewService(m))
	app.Routes("*")
	if err := app.Run(ctx); err != nil {
		t.Fatal(err)
	}

	time.Sleep(10 * time.Millisecond)

	r := jrpc.NewRequest(id.Anyone, testStoragePort)
	//r, _ := jrpc.QueryFlow(id.Anyone, testStoragePort)
	r.Logger(log.New(log.Writer(), "  client ", 0))
	c := client.NewClient(r)

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
		t.Cleanup(func() {
			_ = reader.Close()
		})

		t.Run("idFilter", func(t *testing.T) {
			if f == nil {
				t.Skip()
			}
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

type testStorageModule struct {
	events chan []any
	id.Identity
}

func newTestStorageModule() *testStorageModule {
	m := &testStorageModule{
		events: make(chan []any, 1024),
	}
	var err error
	if m.Identity, err = id.GenerateIdentity(); err != nil {
		panic(err)
	}
	return m
}

func (t *testStorageModule) ReadAll(id data.ID, opts *storage.OpenOpts) (b []byte, err error) {
	if opts.Virtual {
		err = errors.New("test error")
		t.notify("ReadAll", err)
		return
	}
	b = []byte("hello!")
	t.notify("ReadAll", id, opts, b)
	return
}

func (t *testStorageModule) Open(dataID data.ID, opts *storage.OpenOpts) (storage.Reader, error) {
	t.notify("Open", dataID, opts)
	return &testStorageReader{t, opts}, nil
}

type testStorageReader struct {
	*testStorageModule
	opts *storage.OpenOpts
}

func (t *testStorageReader) Read(p []byte) (n int, err error) {
	if t.opts != nil && t.opts.IdentityFilter != nil {
		f := t.opts.IdentityFilter(t.Identity)
		log.Println(f)
	}
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
	return &storageWriter{t, opts}, nil
}

type storageWriter struct {
	*testStorageModule
	opts *storage.CreateOpts
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
