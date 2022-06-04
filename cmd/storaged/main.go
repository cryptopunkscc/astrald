package main

// storage daemon example
//
// Usage: ./starged <node0> <node1> ...
//
// Run a local storage node and use listed nodes for remote queries.

import (
	"errors"
	"fmt"
	astral "github.com/cryptopunkscc/astrald/mod/apphost/client"
	_store "github.com/cryptopunkscc/astrald/proto/store"
	"io"
	"log"
	"os"
	"path/filepath"
)

type Server struct {
	dataDir string
}

func (server *Server) Run() error {
	log.Println("storaged starting...")

	port, err := astral.Register("storage")
	if err != nil {
		return err
	}

	log.Println("data dir:", server.dataDir)

	store := NewMetaStore(server.dataDir, os.Args[1:])

	for conn := range port.AcceptAll() {
		conn := conn
		go func() {
			if err := _store.Serve(conn, store); err != nil {
				if !errors.Is(err, io.EOF) {
					log.Println("error:", err)
				}
			}
			conn.Close()
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
