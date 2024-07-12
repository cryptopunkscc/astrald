package node

import "context"

type ModuleEngine interface {
	Find(name string) Module
	Loaded() []Module
}

type Module interface {
	Run(context.Context) error
}
