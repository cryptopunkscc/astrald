package main

import (
	"fmt"
	"os"
	_debug "runtime/debug"
)

func printVersion() int {
	if info, ok := _debug.ReadBuildInfo(); ok {
		for _, setting := range info.Settings {
			if setting.Key == "vcs.revision" {
				fmt.Println("commit", setting.Value)
				os.Exit(0)
			}
		}
	}
	fmt.Println("no version info available")

	return 0
}
