package media

import (
	"fmt"
	"strings"

	"github.com/cryptopunkscc/astrald/mod/objects"
	"gorm.io/gorm"
)

type tagHandler = func(*gorm.DB) *gorm.DB

// fieldDef defines how a tag field maps to SQL conditions.
type fieldDef struct {
	include string           // SQL template for include, e.g. "LOWER(artist) LIKE ?"
	exclude string           // SQL template for exclude
	arg     func(string) any // transforms the tag value into a SQL argument
}

var audioFields = map[string]fieldDef{
	"artist": {include: "LOWER(artist) LIKE ?", exclude: "LOWER(artist) NOT LIKE ?", arg: likeArg},
	"album":  {include: "LOWER(album) LIKE ?", exclude: "LOWER(album) NOT LIKE ?", arg: likeArg},
	"title":  {include: "LOWER(title) LIKE ?", exclude: "LOWER(title) NOT LIKE ?", arg: likeArg},
	"genre":  {include: "LOWER(genre) LIKE ?", exclude: "LOWER(genre) NOT LIKE ?", arg: likeArg},
	"year":   {include: "year = ?", exclude: "year != ?", arg: exactArg},
}

// textFields defines which fields are searched by free text, in order.
var textFields = []string{"artist", "title", "album"}

func likeArg(v string) any  { return "%" + v + "%" }
func exactArg(v string) any { return v }

type audioQueryBuilder struct {
	db       *gorm.DB
	includes map[string][]string // field -> values; same-field values are OR'd, fields AND'd
	excludes []tagHandler        // each exclude is AND'd
}

func newAudioQuery(db *gorm.DB) *audioQueryBuilder {
	return &audioQueryBuilder{
		db:       db.Model(&dbAudio{}),
		includes: make(map[string][]string),
	}
}

func (b *audioQueryBuilder) Tag(tag objects.QueryTag) *audioQueryBuilder {
	f, ok := audioFields[string(tag.Name)]
	if !ok || tag.Mod == objects.TagModOptional {
		return b
	}
	val := strings.ToLower(string(tag.Value))
	if tag.Mod == objects.TagModExclude {
		b.excludes = append(b.excludes, func(db *gorm.DB) *gorm.DB {
			return db.Where(f.exclude, f.arg(val))
		})
	} else {
		b.includes[string(tag.Name)] = append(b.includes[string(tag.Name)], val)
	}
	return b
}

func (b *audioQueryBuilder) Text(q string) *audioQueryBuilder {
	q = strings.TrimSpace(strings.ToLower(q))
	if q == "" {
		return b
	}

	// only search fields not already constrained by an include tag
	var conds []string
	for _, name := range textFields {
		if _, constrained := b.includes[name]; !constrained {
			conds = append(conds, fmt.Sprintf("LOWER(%s) LIKE ?", name))
		}
	}
	if len(conds) == 0 {
		return b
	}

	args := make([]any, len(conds))
	for i := range conds {
		args[i] = "%" + q + "%"
	}

	b.excludes = append(b.excludes, func(db *gorm.DB) *gorm.DB {
		return db.Where(strings.Join(conds, " OR "), args...)
	})
	return b
}

func (b *audioQueryBuilder) Find() ([]*dbAudio, error) {
	db := b.db

	// OR values within the same field, AND across fields
	for fieldName, values := range b.includes {
		f := audioFields[fieldName]
		conds := make([]string, len(values))
		args := make([]any, len(values))
		for i, v := range values {
			conds[i] = f.include
			args[i] = f.arg(v)
		}
		db = db.Where(strings.Join(conds, " OR "), args...)
	}

	// AND all excludes and text scope
	db = db.Scopes(b.excludes...)

	var rows []*dbAudio
	return rows, db.Find(&rows).Error
}
