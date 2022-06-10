package main

import (
	"fmt"
	"github.com/cryptopunkscc/astrald/data"
	astral "github.com/cryptopunkscc/astrald/mod/apphost/client"
	_store "github.com/cryptopunkscc/astrald/proto/store"
	"io"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		panic("missing args")
	}

	switch os.Args[1] {
	case "r", "read":
		read(os.Args[2])
	case "c", "create":
		create()
	case "i", "identify":
		identify(os.Args[2])
	}
}

func read(id string) {
	blockID, err := data.Parse(id)
	if err != nil {
		panic(err)
	}

	conn, err := astral.DialName("localnode", "storage")
	if err != nil {
		panic(err)
	}

	store := _store.Bind(conn)

	block, err := store.Open(blockID, _store.OpenRemote)
	if err != nil {
		panic(err)
	}

	if _, err := io.Copy(os.Stdout, block); err != nil {
		fmt.Println("copy error:", err)
	}

	if err := block.End(); err != nil {
		fmt.Println("close error:", err)
	}

	conn.Close()
}

func create() {
	conn, err := astral.DialName("localnode", "storage")
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	store := _store.Bind(conn)

	block, tempID, err := store.Create(0)
	if err != nil {
		panic(err)
	}

	fmt.Println("tempID", tempID)

	if _, err := io.Copy(block, os.Stdin); err != nil {
		fmt.Println("copy error:", err)
	}

	id, err := block.Finalize()

	if err != nil {
		fmt.Println("error:", err)
		return
	}

	fmt.Println("stored", id.String())
}

func identify(filename string) {
	file, err := os.Open(filename)
	if err != nil {
		panic(err)
	}

	id, err := data.ResolveAll(file)
	if err != nil {
		panic(err)
	}

	fmt.Println(id.String())
}
