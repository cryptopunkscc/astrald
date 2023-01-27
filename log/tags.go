package log

import "strings"

var tagLevels map[string]int
var tagColors map[string]string

func getTagLevel(tag string) int {
	if l, ok := tagLevels[tag]; ok {
		return l
	}
	return Level
}

func SetTagLevel(tag string, level int) {
	tagLevels[tag] = level
}

func SetTagColor(tag string, color string) {
	var c string
	switch strings.ToLower(color) {
	case "red":
		c = Red()
	case "green":
		c = Green()
	case "yellow":
		c = Yellow()
	case "blue":
		c = Blue()
	case "purple":
		c = Purple()
	case "cyan":
		c = Cyan()
	case "gray":
		c = Gray()
	case "white":
		c = White()
	}
	tagColors[tag] = c
}

func init() {
	tagLevels = make(map[string]int)
	tagColors = make(map[string]string)
}
