package main

import (
	"encoding/hex"
	"flag"
	"fmt"
	"github.com/cryptopunkscc/astrald/data"
	"io"
	"os"
)

func main() {
	var verbose bool

	flag.BoolVar(&verbose, "v", false, "be verbose")
	flag.Parse()

	args := flag.Args()

	if len(args) < 1 {
		fmt.Fprintf(os.Stderr, "usage: %s <file>\n", os.Args[0])
		os.Exit(1)
	}

	var input io.Reader
	var err error

	if args[0] == "-" {
		input = os.Stdin
	} else {
		input, err = os.Open(args[0])
		if err != nil {
			fmt.Fprintln(os.Stderr, "error:", err)
			os.Exit(2)
		}
	}

	dataID, err := data.ResolveAll(input)
	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(3)
	}

	if verbose {
		fmt.Println("id:  ", dataID.String())
		fmt.Println("size:", dataID.Size)
		fmt.Println("hash:", hex.EncodeToString(dataID.Hash[:]))
	} else {
		fmt.Println(dataID.String())
	}
}
