package zip

import "github.com/cryptopunkscc/astrald/data"

type ArchiveDesc struct {
	Files []ArchiveFile
}

func (ArchiveDesc) Type() string {
	return "mod.zip.archive"
}

type ArchiveFile struct {
	DataID data.ID
	Path   string
}

type MemberDesc struct {
	Memberships []Membership
}

func (MemberDesc) Type() string {
	return "mod.zip.member"
}

type Membership struct {
	ZipID data.ID
	Path  string
}
