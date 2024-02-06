package media

import "time"

const ModuleName = "media"
const DBPrefix = "media__"

type Module interface {
}

type Info struct {
	Type     string // audio|video
	Title    string
	Artist   string
	Album    string
	Genre    string
	Duration time.Duration
}
