package term

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/object"
)

const HighlightColor = "yellow"
const ErrorColor = "red"
const DefaultColor = "default"

const maxTranslateDepth = 3

var DefaultTypeMap = TypeMap{}

func init() {
	// astral.Identity
	SetTranslateFunc(func(object *astral.Identity) astral.Object {
		return &ColorString{
			Color: HighlightColor,
			Text:  astral.String32(Render(object, nil, true)),
		}
	})

	// astral.ErrorMessage
	SetTranslateFunc(func(object *astral.ErrorMessage) astral.Object {
		return &ColorString{
			Color: ErrorColor,
			Text:  astral.String32(Render(object, nil, true)),
		}
	})

	// astral.Nonce
	SetTranslateFunc(func(object *astral.Nonce) astral.Object {
		return &ColorString{
			Color: HighlightColor,
			Text:  astral.String32(Render(object, nil, true)),
		}
	})

	// object.ID
	SetTranslateFunc(func(object *object.ID) astral.Object {
		return &ColorString{
			Color: "blue",
			Text:  astral.String32(Render(object, nil, true)),
		}
	})
}
