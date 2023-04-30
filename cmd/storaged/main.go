package main

// storage daemon example
//
// Usage: ./starged <node0> <node1> ...
//
// Run a local storage node and use listed nodes for remote queries.

import (
	"errors"
	"fmt"
	"github.com/cryptopunkscc/astrald/lib/astral"
	_log "github.com/cryptopunkscc/astrald/log"
	_store "github.com/cryptopunkscc/astrald/proto/store"
	"io"
	"os"
	"path/filepath"
)

type Server struct {
	dataDir string
}

var log = _log.Tag("storaged")

func (server *Server) Run() error {
	log.Log("storaged starting...")

	port, err := astral.Listen("storage")
	if err != nil {
		return err
	}

	log.Log("data dir: %s", server.dataDir)

	store := NewMetaStore(server.dataDir, os.Args[1:])

	for query := range port.QueryCh() {
		info, _ := astral.NodeInfo(query.RemoteIdentity())
		log.Log("%s connected", info.Name)

		conn, err := query.Accept()
		if err != nil {
			continue
		}
		go func() {
			defer conn.Close()

			err := _store.Serve(conn, NewAuthStore(conn, store))
			if (err != nil) && !errors.Is(err, io.EOF) {
				log.Error("serve: %s", err)
			}
			log.Log("%s disconnected", info.Name)
		}()
	}

	return nil
}

func main() {
	var dataDir = "./storage"

	if home, err := os.UserHomeDir(); err == nil {
		dataDir = filepath.Join(home, ".config/astrald/storage")
	}

	if err := os.MkdirAll(dataDir, 0700); err != nil {
		fmt.Println("init error:", err)
		os.Exit(1)
	}

	server := &Server{dataDir: dataDir}

	if err := server.Run(); err != nil {
		fmt.Println("server error:", err)
		os.Exit(2)
	}
}
