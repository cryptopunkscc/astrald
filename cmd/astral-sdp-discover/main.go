package main

import (
	"fmt"
	"github.com/cryptopunkscc/astrald/lib/astral"
	"os"
)

func main() {
	var args = os.Args[1:]
	var target = "localnode"

	if len(args) >= 1 {
		target = args[0]
	}

	identity, err := astral.Resolve(target)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %s\n", err)
		os.Exit(1)
	}

	info, err := astral.GetNodeInfo(identity)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %s\n", err)
		os.Exit(1)
	}

	fmt.Printf("Discovering services on %s...\n", info.Name)

	services, err := astral.Client.Discovery().Discover(identity)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %s\n", err)
		os.Exit(2)
	}

	var f = "%-20s %-25s %-25s %s\n"

	fmt.Printf(f, "IDENTITY", "SERVICE", "TYPE", "EXTRA")
	for _, s := range services {
		var name = s.Identity.Fingerprint()

		if info, err := astral.GetNodeInfo(s.Identity); err == nil {
			name = info.Name
		}

		fmt.Printf(f, name, s.Name, s.Type, string(s.Extra))
	}
}
