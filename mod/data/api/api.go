package data

import (
	"github.com/cryptopunkscc/astrald/data"
	"github.com/cryptopunkscc/astrald/node/events"
	"github.com/cryptopunkscc/astrald/node/modules"
	"io"
	"time"
)

const ModuleName = "data"

type API interface {
	Events() *events.Queue
	FindByType(typ string, since time.Time) ([]TypeInfo, error)
	OpenADC0(data.ID) (*ADC0Header, io.ReadCloser, error)
	SetLabel(data.ID, string)
	GetLabel(data.ID) string
}

type TypeInfo struct {
	ID        data.ID
	IndexedAt time.Time
	Header    string
	Type      string
}

type EventDataIndexed TypeInfo

func Load(node modules.Node) (API, error) {
	api, ok := node.Modules().Find(ModuleName).(API)
	if !ok {
		return nil, modules.ErrNotFound
	}
	return api, nil
}
