package cslq

const (
	Port = "storage."
)

const (
	ReadAll = byte(iota) + 1
	Put

	AddOpener
	RemoveOpener
	AddCreator
	RemoveCreator
	AddPurger
	RemovePurger

	OpenerOpen
	CreatorCreate
	PurgerPurge
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
