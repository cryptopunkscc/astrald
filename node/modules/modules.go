package modules

type Modules interface {
	Find(name string) Module
	Loaded() []Module
}

func Load[M any](node Node, name string) (M, error) {
	mod, ok := node.Modules().Find(name).(M)
	if !ok {
		return mod, ModuleUnavailable(name)
	}
	return mod, nil
}
