package zip

import "github.com/cryptopunkscc/astrald/data"

const ArchiveDescriptorType = "mod.zip.archive"

type ArchiveDescriptor struct {
	Files []ArchiveFile
}

type ArchiveFile struct {
	DataID data.ID
	Path   string
}

const MemberDescriptorType = "mod.zip.member"

type MemberDescriptor struct {
	Memberships []Membership
}

type Membership struct {
	ZipID data.ID
	Path  string
}
