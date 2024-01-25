package webdata

import (
	"context"
	"embed"
	"errors"
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/log"
	"github.com/cryptopunkscc/astrald/mod/data"
	"github.com/cryptopunkscc/astrald/mod/index"
	"github.com/cryptopunkscc/astrald/mod/shares"
	"github.com/cryptopunkscc/astrald/mod/storage"
	"github.com/cryptopunkscc/astrald/node"
	"github.com/cryptopunkscc/astrald/resources"
	"net/http"
)

//go:embed res/*
var res embed.FS

type Module struct {
	config   Config
	node     node.Node
	log      *log.Logger
	assets   resources.Resources
	identity id.Identity
	mux      *http.ServeMux

	storage storage.Module
	shares  shares.Module
	index   index.Module
	data    data.Module

	rootHandler  *RootHandler
	dataHandler  *DataHandler
	indexHandler *IndexHandler
}

func (mod *Module) Run(ctx context.Context) error { // Create a new ServeMux
	var addr = mod.config.Listen

	if addr == "" {
		return errors.New("listen addr not configured")
	}

	var server = &http.Server{
		Addr:    addr,
		Handler: mod.mux,
	}

	go func() {
		mod.log.Info("listen %v as %v", addr, mod.identity)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			mod.log.Error("listen error: %v", err)
		}
	}()

	<-ctx.Done()
	server.Close()

	return nil
}
