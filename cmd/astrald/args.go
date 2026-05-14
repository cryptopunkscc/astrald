package main

import (
	"flag"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

const nodeDirName = "astrald"

type Args struct {
	Root    string
	Ghost   bool
	Version bool
}

func parseArgs() *Args {
	var root string
	var ghost, version bool

	flag.StringVar(&root, "root", "", "set node's root directory (config goes to <root>/config, data to <root>/data)")
	flag.BoolVar(&ghost, "g", false, "enable ghost mode")
	flag.BoolVar(&version, "v", false, "show version")
	flag.Parse()

	args := &Args{
		Ghost:   ghost,
		Version: version,
	}

	if root != "" {
		args.Root = root
	}

	if strings.HasPrefix(args.Root, "~/") {
		if homeDir, err := os.UserHomeDir(); err == nil {
			args.Root = filepath.Join(homeDir, args.Root[2:])
		}
	}

	return args
}

func userDataDir() (string, error) {
	// On macOS, there is no XDG-style split — both config and data
	// live under ~/Library/Application Support. Return the same
	// directory as os.UserConfigDir() so they coexist.
	if runtime.GOOS == "darwin" {
		return os.UserConfigDir()
	}

	if dir := os.Getenv("XDG_DATA_HOME"); dir != "" {
		return dir, nil
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	return filepath.Join(home, ".local", "share"), nil
}

func defaultConfigRoot() string {
	cfgDir, err := os.UserConfigDir()
	if err != nil {
		return nodeDirName
	}

	return filepath.Join(cfgDir, nodeDirName)
}

func defaultDataRoot() string {
	dataDir, err := userDataDir()
	if err != nil {
		return filepath.Join(os.Getenv("HOME"), nodeDirName)
	}

	return filepath.Join(dataDir, nodeDirName)
}
