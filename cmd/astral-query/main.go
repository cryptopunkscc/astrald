package main

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/lib/arl"
	"github.com/cryptopunkscc/astrald/lib/astrald"
	"github.com/cryptopunkscc/astrald/lib/query"
	dircli "github.com/cryptopunkscc/astrald/mod/dir/client"
)

const (
	EnvDefaultTarget = "ASTRAL_DEFAULT_TARGET"
	EnvDefaultIn     = "ASTRAL_DEFAULT_INPUT_FORMAT"
	EnvDefaultOut    = "ASTRAL_DEFAULT_OUTPUT_FORMAT"
)

func main() {
	// show help
	if len(os.Args) < 2 {
		fmt.Printf("usage: %s <query> [-arg <val>]...\n", os.Args[0])
		return
	}

	defaultIn := os.Getenv(EnvDefaultIn)
	defaultOut := os.Getenv(EnvDefaultOut)

	var err error
	var callerID, targetID *astral.Identity

	// split the argument into parts
	caller, target, method := arl.Split(os.Args[1])

	// create new astral context
	var ctx = astrald.NewContext()

	// parse the caller
	if len(caller) > 0 {
		callerID, err = dircli.ResolveIdentity(ctx, caller)
		if err != nil {
			fatal("error: %v\n", err)
		}
	}

	// set default target if none given
	if len(target) == 0 {
		target = os.Getenv(EnvDefaultTarget)
	}

	// parse the target
	if len(target) > 0 {
		targetID, err = dircli.ResolveIdentity(ctx, target)
		if err != nil {
			fatal("error: %v\n", err)
		}
	}

	// parse the arguments
	var osArgs = os.Args[2:]
	var args = map[string]string{}
	for len(osArgs) >= 2 {
		key := osArgs[0]
		val := osArgs[1]
		if !strings.HasPrefix(key, "-") || len(key) < 2 {
			fatal("error: unexpected argument %s\n", key)
		}
		args[key[1:]] = val
		osArgs = osArgs[2:]
	}

	if len(osArgs) == 1 {
		args[query.DefaultArgKey] = osArgs[0]
	}

	// set default input/output formats
	if defaultIn != "" && args["in"] == "" {
		args["in"] = defaultIn
	}
	if defaultOut != "" && args["out"] == "" {
		args["out"] = defaultOut
	}

	// route the query
	conn, err := astrald.RouteQuery(ctx, query.New(callerID, targetID, method, args))
	if err != nil {
		fatal("error: %v\n", err)
	}

	// join conn with the terminal
	go func() {
		io.Copy(conn, os.Stdin)
	}()
	io.Copy(os.Stdout, conn)
}

func fatal(f string, v ...any) {
	fmt.Fprintf(os.Stderr, f, v...)
	os.Exit(1)
}
