package theme

import . "github.com/cryptopunkscc/astrald/mod/log/styles"

var Normal = White

// Theme colors
var (
	Primary                         = Orange.Bri(1.1)
	Secondary, Tertiary, Quaternary = Primary.Tetrad()
)

const (
	Most  = 1.8
	More  = 1.44
	Less  = 0.84
	Least = 0.62
)

var (
	Ack      = Green.Bri(Less)
	True     = Green.Bri(More)
	False    = Red.Bri(More)
	EOS      = White.Bri(Least)
	Error    = Red
	Identity = Primary
	Level    = White.Bri(Least)
	Nonce    = Quaternary.Bri(Least)
	Nil      = Red.Bri(Less)
	ObjectID = Tertiary
	Op       = Yellow.Bri(Most)
	Size     = Primary
	Time     = Normal.Bri(Least)
	Type     = Tertiary.Bri(Most)
)
