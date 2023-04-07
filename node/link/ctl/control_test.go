package ctl

import (
	"github.com/cryptopunkscc/astrald/streams"
	"reflect"
	"sync"
	"testing"
)

func TestControl(t *testing.T) {
	w, r := streams.Pipe()

	var wg = sync.WaitGroup{}

	// writer
	wg.Add(1)
	go func() {
		defer wg.Done()

		ctl := New(w)
		if err := ctl.WriteDrop(69); err != nil {
			panic(err)
		}
		if err := ctl.WriteQuery("test", 420); err != nil {
			panic(err)
		}
		if err := ctl.WriteClose(); err != nil {
			panic(err)
		}
	}()

	// reader
	wg.Add(1)
	go func() {
		defer wg.Done()
		var ctl = New(r)

		// read DropMessage
		msg, err := ctl.ReadMessage()
		if err != nil {
			t.Error("read error:", err)
			return
		}

		if drop, ok := msg.(DropMessage); ok {
			if drop.Port() != 69 {
				t.Error("expected 69, got", drop.Port())
			}
		} else {
			t.Error("expected DropMessage, got", reflect.TypeOf(msg))
		}

		// read QueryMessage
		msg, err = ctl.ReadMessage()
		if err != nil {
			t.Error("read error:", err)
			return
		}

		if query, ok := msg.(QueryMessage); ok {
			if query.Port() != 420 {
				t.Error("expected 420, got", query.Port())
			}
			if query.Query() != "test" {
				t.Error("exptected test, got", query.Query())
			}
		} else {
			t.Error("expected QueryMessage, got", reflect.TypeOf(msg))
		}

		// read CloseMessage
		msg, err = ctl.ReadMessage()
		if err != nil {
			t.Error("read error:", err)
			return
		}

		if _, ok := msg.(CloseMessage); !ok {
			t.Error("expected CloseMessage, got", reflect.TypeOf(msg))
		}
	}()

	wg.Wait()
}
