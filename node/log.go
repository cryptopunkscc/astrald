package node

import (
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/log"
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
