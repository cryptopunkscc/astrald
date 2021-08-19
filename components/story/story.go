package story

import (
	"bytes"
	"github.com/cryptopunkscc/astrald/components/fid"
)

const (
	BeginMagicBytes  = 0
	BeginTimestamp   = 4
	BeginTypeSize    = 12
	BeginAuthorSize  = 13
	BeginRefsCount   = 14
	BeginDataSize    = 15
	BeginDynamicData = 17
)

const (
	LenMagicBytes = 4
	LenTimestamp  = 8
	LenTypeSize   = 1
	LenAuthorSize = 1
	LenRefsCount  = 1
	LenDataSize   = 2
)

const HeaderSize = LenMagicBytes +
	LenTimestamp +
	LenTypeSize +
	LenAuthorSize +
	LenRefsCount +
	LenDataSize

var MagicBytes = [LenMagicBytes]byte{0, 4, 2, 0}

type Story struct {
	timestamp uint64
	typ       []byte
	author    []byte
	refs      []fid.ID
	data      []byte
}

func NewStory(
	timestamp int64,
	typ string,
	author string,
	refs []fid.ID,
	data []byte,
) *Story {
	return &Story{
		timestamp: uint64(timestamp),
		typ:       bytes.NewBufferString(typ).Bytes(),
		author:    bytes.NewBufferString(author).Bytes(),
		refs:      refs,
		data:      data,
	}
}

func (s Story) Author() string {
	return string(s.author)
}
func (s *Story) Timestamp() int64 {
	return int64(s.timestamp)
}
func (s *Story) Type() string {
	return string(s.typ)
}
func (s *Story) Refs() []fid.ID {
	return s.refs
}
func (s *Story) Data() []byte {
	return s.data
}
func (s *Story) SetTimestamp(timestamp int64) *Story {
	s.timestamp = uint64(timestamp)
	return s
}
func (s *Story) SetType(typ string) *Story {
	s.typ = bytes.NewBufferString(typ).Bytes()
	return s
}
func (s *Story) SetAuthor(author string) *Story {
	s.author = bytes.NewBufferString(author).Bytes()
	return s
}
func (s *Story) SetRefs(refs []fid.ID) *Story {
	s.refs = refs
	return s
}
func (s *Story) SetData(data []byte) *Story {
	s.data = data
	return s
}
