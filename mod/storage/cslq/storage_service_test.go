package cslq

import (
	"context"
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/data"
	"github.com/cryptopunkscc/astrald/lib/astral"
	"github.com/cryptopunkscc/astrald/mod/storage"
	client "github.com/cryptopunkscc/astrald/mod/storage/cslq/client"
	"github.com/go-playground/assert/v2"
	"path"
	"testing"
	"time"
)

const testStoragePort = client.Port + "test"

func TestNewStorageService(tt *testing.T) {
	// start service
	t := newTestStorageModule()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	service := NewStorageService(nil, t)
	service.port = testStoragePort
	if err := service.run(ctx); err != nil {
		tt.Error(err)
	}

	// prepare client
	c := client.NewClient(id.Identity{}).Port(testStoragePort)

	// run tests
	t.testOpen(tt, c, nil)
	t.testCreate(tt, c)
	t.testPurge(tt, c)
	t.testPut(tt, c)
	t.testReadAll(tt, c)
	t.testAddOpener(tt, ctx, c)
	t.testAddCreator(tt, ctx, c)
}

func (t *testStorageModule) testReadAll(
	tt *testing.T,
	s storage.Module,
) {
	tt.Run("ReadAll", func(tt *testing.T) {
		dataID := data.ID{Size: 1}
		opts := &storage.OpenOpts{Offset: 1}
		bytes, err := s.ReadAll(dataID, opts)
		t.verifyEq(tt, err, dataID, opts, bytes)
	})
}

func (t *testStorageModule) testPut(
	tt *testing.T,
	s storage.Module,
) {
	tt.Run("Put", func(tt *testing.T) {
		bytes := []byte{1}
		opts := &storage.CreateOpts{Alloc: 1}
		put, err := s.Put(bytes, opts)
		t.verifyEq(tt, err, bytes, opts, put)
	})
}

func (t *testStorageModule) testPurge(
	tt *testing.T,
	s storage.Purger,
) {
	tt.Run("Purge", func(tt *testing.T) {
		dataID := data.ID{Size: 1}
		opts := &storage.PurgeOpts{}
		l, err := s.Purge(dataID, opts)
		t.verifyEq(tt, err, dataID, opts, l)
	})
}

func (t *testStorageModule) testCreate(
	tt *testing.T,
	s storage.Creator,
) {
	tt.Run("Create", func(tt *testing.T) {
		opts := &storage.CreateOpts{Alloc: 1}
		writer, err := s.Create(opts)
		t.verifyEq(tt, err, opts)

		tt.Run("Write", func(tt *testing.T) {
			b := make([]byte, 16)
			l, err := writer.Write(b)
			t.verifyEq(tt, err, b, l)
		})

		tt.Run("Commit", func(tt *testing.T) {
			d, err := writer.Commit()
			t.verifyEq(tt, err, d)
		})

		tt.Run("Discard", func(tt *testing.T) {
			err := writer.Discard()
			t.verifyEq(tt, err)
		})
	})
}

func (t *testStorageModule) testOpen(
	tt *testing.T,
	s storage.Opener,
	f id.Filter,
) {
	tt.Run("Open", func(tt *testing.T) {
		dataID := data.ID{Size: 1}
		opts := &storage.OpenOpts{Offset: 1, IdentityFilter: f}
		reader, err := s.Open(dataID, opts)

		t.verify(tt, err, func(left []any, right []any) {
			left[2] = right[2]
			assert.Equal(tt, left, right)
			if f != nil {
				f = right[2].(*storage.OpenOpts).IdentityFilter
			}
		}, dataID, opts)
		tt.Cleanup(func() { _ = reader.Close() })

		if f != nil {
			tt.Run("idFilter", func(tt *testing.T) {
				b := f(id.Anyone)
				t.verifyEq(tt, nil, id.Anyone, b)
			})
		}

		tt.Run("Info", func(tt *testing.T) {
			info := reader.Info()
			t.verifyEq(tt, err, info)
		})

		tt.Run("Seek", func(tt *testing.T) {
			offset := int64(1)
			whence := 1
			seek, err := reader.Seek(offset, whence)
			t.verifyEq(tt, err, offset, whence, seek)
		})

		tt.Run("Read", func(tt *testing.T) {
			b := make([]byte, 16)
			read, err := reader.Read(b)
			t.verifyEq(tt, err, b, read)
		})
	})
}

func (t *testStorageModule) testAddCreator(
	tt *testing.T,
	ctx context.Context,
	s storage.Module,
) {
	tt.Run("AddCreator", func(tt *testing.T) {
		priority := 1
		name := "creator"
		port := "creator.service.test"
		cs := client.NewCreatorService(t).Port(port)
		if err := client.RegisterHandler(ctx, cs); err != nil {
			return
		}
		err := s.AddCreator(name, cs, priority)
		var creator storage.Creator
		t.verify(tt, err, func(left []any, right []any) {
			left[2] = right[2]
			assert.Equal(tt, left, right)
			creator = right[2].(storage.Creator)
		}, name, port, priority)

		t.testCreate(tt, creator)
	})
}

func (t *testStorageModule) testAddOpener(
	tt *testing.T,
	ctx context.Context,
	s storage.Module,
) {
	tt.Run("AddOpener", func(tt *testing.T) {
		priority := 1
		name := "opener"
		port := "opener.service.test"
		os := client.NewOpenerService(t, id.Anyone).Port(port)
		if err := client.RegisterHandler(ctx, os); err != nil {
			return
		}
		err := s.AddOpener(name, os, priority)
		var opener storage.Opener
		t.verify(tt, err, func(left []any, right []any) {
			left[2] = right[2]
			assert.Equal(tt, left, right)
			opener = right[2].(storage.Opener)
		}, name, port, priority)

		f := func(identity id.Identity) (b bool) {
			b = true
			t.notify("idFilter", identity, b)
			return
		}
		t.testOpen(tt, opener, f)
	})
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
	t.notify("ReadAll", id, opts, b)
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
	name := path.Base(tt.Name())
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

func (s *StorageService) run(ctx context.Context) (err error) {
	l, err := astral.Register(s.port + "*")
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
				conn, err := qd.Accept()
				defer conn.Close()
				if err != nil {
					return
				}
				if err = s.decode(conn, conn.RemoteIdentity(), conn.Query()); err != nil {
					return
				}
			}(qd)
		}
	}()
	return
}
