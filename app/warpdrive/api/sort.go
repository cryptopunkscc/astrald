package api

type ByTimestamp []Offer

func (a ByTimestamp) Len() int           { return len(a) }
func (a ByTimestamp) Less(i, j int) bool { return a[i].Create < a[j].Create }
func (a ByTimestamp) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
