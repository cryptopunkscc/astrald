package storage

import "github.com/cryptopunkscc/astrald/auth/id"

type DataSource struct {
	Identity id.Identity
	Service  string
}

func (mod *Module) AddDataSource(source *DataSource) {
	mod.dataSourcesMu.Lock()
	defer mod.dataSourcesMu.Unlock()

	mod.dataSources[source] = struct{}{}
	mod.log.Info("%s registered source %s", source.Identity, source.Service)
}

func (mod *Module) RemoveDataSource(source *DataSource) {
	mod.dataSourcesMu.Lock()
	defer mod.dataSourcesMu.Unlock()

	delete(mod.dataSources, source)
	mod.log.Logv(1, "%s unregistered source %s", source.Identity, source.Service)
}

func (mod *Module) DataSources() []*DataSource {
	mod.dataSourcesMu.Lock()
	defer mod.dataSourcesMu.Unlock()

	var list = make([]*DataSource, 0, len(mod.dataSources))
	for source := range mod.dataSources {
		list = append(list, source)
	}
	return list
}
