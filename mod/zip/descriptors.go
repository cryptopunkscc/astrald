package zip

import "github.com/cryptopunkscc/astrald/object"

type ArchiveDesc struct {
	Files []ArchiveFile
}

func (ArchiveDesc) Type() string {
	return "mod.zip.archive"
}

type ArchiveFile struct {
	ObjectID object.ID
	Path     string
}

type MemberDesc struct {
	Memberships []Membership
}

func (MemberDesc) Type() string {
	return "mod.zip.member"
}

type Membership struct {
	ZipID object.ID
	Path  string
}
