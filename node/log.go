package node

import (
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/data"
	"github.com/cryptopunkscc/astrald/log"
	"strings"
	"time"
)

// aliasString is a type used to force formatting of aliases
type aliasString string

func (node *CoreNode) pushLogFormatters() {
	// string format
	node.log.PushFormatFunc(func(v any) ([]log.Op, bool) {
		s, ok := v.(string)
		if !ok {
			return nil, false
		}
		return []log.Op{
			log.OpColor{Color: log.Yellow},
			log.OpText{Text: s},
			log.OpReset{},
		}, true
	})

	// aliasString format
	node.log.PushFormatFunc(func(v any) ([]log.Op, bool) {
		s, ok := v.(aliasString)
		if !ok {
			return nil, false
		}
		return []log.Op{
			log.OpColor{Color: log.Green},
			log.OpText{Text: string(s)},
			log.OpReset{},
		}, true
	})

	// error format
	node.log.PushFormatFunc(func(v any) ([]log.Op, bool) {
		s, ok := v.(error)
		if !ok {
			return nil, false
		}
		return []log.Op{
			log.OpColor{Color: log.Red},
			log.OpText{Text: s.Error()},
			log.OpReset{},
		}, true
	})

	// time.Duration format
	node.log.PushFormatFunc(func(v any) ([]log.Op, bool) {
		s, ok := v.(time.Duration)
		if !ok {
			return nil, false
		}
		return []log.Op{
			log.OpColor{Color: log.Magenta},
			log.OpText{Text: s.String()},
			log.OpReset{},
		}, true
	})

	// id.Identity format
	node.log.PushFormatFunc(func(v any) ([]log.Op, bool) {
		identity, ok := v.(id.Identity)
		if !ok {
			return nil, false
		}

		var color = log.Cyan

		if node.identity.IsEqual(identity) {
			color = log.Green
		}

		var name = node.Resolver().DisplayName(identity)

		return []log.Op{
			log.OpColor{Color: color},
			log.OpText{Text: name},
			log.OpReset{},
		}, true
	})

	node.log.PushFormatFunc(func(v any) ([]log.Op, bool) {
		dataID, ok := v.(data.ID)
		if !ok {
			return nil, false
		}

		var ops []log.Op
		var s = dataID.String()

		if strings.HasPrefix(s, "id1") {
			s = s[3:]
			ops = append(ops,
				log.OpColor{Color: log.Blue},
				log.OpText{Text: "id1"},
				log.OpReset{},
			)
		}

		ops = append(ops,
			log.OpColor{Color: log.BrightBlue},
			log.OpText{Text: s},
			log.OpReset{},
		)

		return ops, true
	})

	// nil format
	node.log.PushFormatFunc(func(v any) ([]log.Op, bool) {
		if v != nil {
			return nil, false
		}
		return []log.Op{
			log.OpColor{Color: log.Red},
			log.OpText{Text: "nil"},
			log.OpReset{},
		}, true
	})
}
