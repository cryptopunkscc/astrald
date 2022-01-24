package core

import "log"

func (c *core) SetLogger(logger *log.Logger) {
	c.Logger = logger
}
