package proto

const (
	Port = "storage."
)

const (
	WriterWrite = byte(iota) + 1
	WriterCommit
	WriterDiscard
)

const (
	ReaderRead = byte(iota) + 1
	ReaderSeek
	ReaderInfo
	ReaderClose
)
