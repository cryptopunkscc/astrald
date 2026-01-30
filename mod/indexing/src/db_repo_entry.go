package indexing

import (
	"time"

	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/mod/indexing"
)

type dbRepoEntry struct {
	Repo      string           `gorm:"index;primaryKey;uniqueIndex:idx_repo_version"`
	ObjectID  *astral.ObjectID `gorm:"index;primaryKey"`
	Version   int              `gorm:"index;uniqueIndex:idx_repo_version"`
	Exist     bool             `gorm:"index"`
	CreatedAt time.Time
	UpdatedAt time.Time
}

func (dbRepoEntry) TableName() string { return indexing.DBPrefix + "repo_entries" }
