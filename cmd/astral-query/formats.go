package main

import (
	"os"

	"golang.org/x/term"
)

const EnvDefaultIn = "ASTRAL_DEFAULT_INPUT_FORMAT"
const EnvDefaultOut = "ASTRAL_DEFAULT_OUTPUT_FORMAT"

func inputFormat() string {
	if !term.IsTerminal(int(os.Stdin.Fd())) {
		return ""
	}

	defaultIn := os.Getenv(EnvDefaultIn)
	if defaultIn != "" {
		return defaultIn
	}

	return "text"
}

func outputFormat() string {
	if !term.IsTerminal(int(os.Stdout.Fd())) {
		return ""
	}

	defaultOut := os.Getenv(EnvDefaultOut)
	if defaultOut != "" {
		return defaultOut
	}

	return "render"
}
