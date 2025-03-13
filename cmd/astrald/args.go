package main

import (
	"flag"
	"os"
	"path/filepath"
	"strings"
)

const nodeDirName = "astrald"

type Args struct {
	NodeRoot string
	DBRoot   string
	Ghost    bool
	Version  bool
}

func parseArgs() *Args {
	var args = &Args{
		NodeRoot: defaultRoot(),
	}

	flag.StringVar(&args.NodeRoot, "root", args.NodeRoot, "set node's root directory")
	flag.StringVar(&args.DBRoot, "dbroot", args.NodeRoot, "set database root directory")
	flag.BoolVar(&args.Ghost, "g", false, "enable ghost mode")
	flag.BoolVar(&args.Version, "v", false, "show version")
	flag.Parse()

	if strings.HasPrefix(args.NodeRoot, "~/") {
		if homeDir, err := os.UserHomeDir(); err == nil {
			args.NodeRoot = filepath.Join(homeDir, args.NodeRoot[2:])
		}
	}

	return args
}

func defaultRoot() string {
	cfgDir, err := os.UserConfigDir()
	if err != nil {
		return nodeDirName
	}

	return filepath.Join(cfgDir, "astrald")
}
