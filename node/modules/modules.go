package modules

type Modules interface {
	Find(name string) Module
	Loaded() []Module
}
