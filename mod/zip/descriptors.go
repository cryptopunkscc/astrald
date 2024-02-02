package zip

import "github.com/cryptopunkscc/astrald/data"

type ArchiveDescriptor struct {
	Files []ArchiveFile
}

func (ArchiveDescriptor) InfoType() string {
	return "mod.zip.archive"
}

type ArchiveFile struct {
	DataID data.ID
	Path   string
}

type MemberDescriptor struct {
	Memberships []Membership
}

func (MemberDescriptor) InfoType() string {
	return "mod.zip.member"
}

type Membership struct {
	ZipID data.ID
	Path  string
}
