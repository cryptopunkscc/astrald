package cslq

type String8 string
type String16 string
type String32 string
type String64 string

type Buffer8 []byte
type Buffer16 []byte
type Buffer32 []byte
type Buffer64 []byte

func (s String8) FormatCSLQ() string {
	return "[c]c"
}

func (s String16) FormatCSLQ() string {
	return "[s]c"
}

func (s String32) FormatCSLQ() string {
	return "[l]c"
}

func (s String64) FormatCSLQ() string {
	return "[q]c"
}

func (s Buffer8) FormatCSLQ() string {
	return "[c]c"
}

func (s Buffer16) FormatCSLQ() string {
	return "[s]c"
}

func (s Buffer32) FormatCSLQ() string {
	return "[l]c"
}

func (s Buffer64) FormatCSLQ() string {
	return "[q]c"
}
