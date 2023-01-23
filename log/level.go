package log

var tagLevels map[string]int

func getTagLevel(tag string) int {
	return tagLevels[tag]
}

func SetTagLevel(tag string, level int) {
	tagLevels[tag] = level
}

func init() {
	tagLevels = make(map[string]int)
}
