package webdata

type Config struct {
	Listen   string
	Identity string
}

var defaultConfig = Config{
	Listen: "127.0.0.1:8080",
}
