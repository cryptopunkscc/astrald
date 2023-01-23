package log

var tagLevels map[string]int

func getTagLevel(tag string) int {
	if l, ok := tagLevels[tag]; ok {
		return l
	}
	return LogLevel
}

func SetTagLevel(tag string, level int) {
	tagLevels[tag] = level
}

func init() {
	tagLevels = make(map[string]int)
}
