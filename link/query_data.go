package link

type queryData struct {
	StreamID int
	Query    string
}

func (q queryData) FormatCSLQ() string {
	return "s[c]c"
}
