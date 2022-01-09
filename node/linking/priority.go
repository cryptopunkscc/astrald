package linking

type priorityList []string

var netPriorities = priorityList{"inet", "gw", "bt", "tor"}

func (list priorityList) Priority(item string) int {
	for i, v := range list {
		if v == item {
			return i
		}
	}

	return len(list)
}
