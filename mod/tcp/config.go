package tcp

import "time"

type Config struct {
	DialTimeout     time.Duration
	PublicEndpoints []string
	ListenPort      int
}

var defaultConfig = Config{
	DialTimeout: time.Minute,
	ListenPort:  1791,
}
