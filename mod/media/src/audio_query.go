package media

import (
	"strconv"
	"strings"

	"github.com/cryptopunkscc/astrald/mod/objects"
	"gorm.io/gorm"
)

// knownAudioTags lists the tag names the audio searcher understands.
// Also used as the column whitelist to prevent SQL injection in apply().
var knownAudioTags = []string{"artist", "album", "title", "genre", "year", "format"}

var knownAudioTagSet = func() map[string]bool {
	m := make(map[string]bool, len(knownAudioTags))
	for _, t := range knownAudioTags {
		m[t] = true
	}
	return m
}()

// audioQuery holds parsed search criteria for audio files.
type audioQuery struct {
	text    string              // free-text term (lowercased)
	include map[string][]string // tag → values; OR within group, AND between groups
	exclude map[string][]string // tag → values; AND NOT each group
}

// parseAudioQuery converts a SearchQuery into an audioQuery.
// Required (+) and optional (bare) include tags both add SQL conditions.
// Required (-) and optional (!) exclude tags both add SQL NOT conditions.
func parseAudioQuery(q objects.SearchQuery) *audioQuery {
	aq := &audioQuery{
		text:    strings.ToLower(strings.TrimSpace(string(q.Query))),
		include: make(map[string][]string),
		exclude: make(map[string][]string),
	}
	for _, tag := range q.Tags {
		name := string(tag.Name)
		if !knownAudioTagSet[name] {
			continue
		}
		val := string(tag.Value)
		switch tag.Mod {
		case objects.TagModRequire:
			aq.include[name] = append(aq.include[name], val)
		case objects.TagModExclude:
			aq.exclude[name] = append(aq.exclude[name], val)
			// TagModOptional and TagModOptionalExclude are ignored (no ranking support)
		}
	}
	return aq
}

// apply builds GORM WHERE conditions onto db and returns the modified *gorm.DB.
func (aq *audioQuery) apply(db *gorm.DB) *gorm.DB {
	// free text: OR across artist, title, album
	if aq.text != "" {
		like := "%" + aq.text + "%"
		db = db.Where(
			"LOWER(artist) LIKE ? OR LOWER(title) LIKE ? OR LOWER(album) LIKE ?",
			like, like, like,
		)
	}

	// include groups: AND between tag names, OR within
	for name, vals := range aq.include {
		db = applyTagCondition(db, false, name, vals)
	}

	// exclude groups: AND NOT between tag names, OR within
	for name, vals := range aq.exclude {
		db = applyTagCondition(db, true, name, vals)
	}

	return db
}

// applyTagCondition builds one include or exclude clause for a single tag name.
func applyTagCondition(db *gorm.DB, negate bool, name string, vals []string) *gorm.DB {
	var clauses []string
	var args []interface{}

	for _, v := range vals {
		if name == "year" {
			c, a, ok := buildYearClause(v)
			if !ok {
				continue
			}
			clauses = append(clauses, c)
			args = append(args, a...)
		} else if name == "format" {
			clauses = append(clauses, "LOWER(format) = ?")
			args = append(args, v)
		} else {
			clauses = append(clauses, "LOWER("+name+") LIKE ?")
			args = append(args, "%"+v+"%")
		}
	}

	if len(clauses) == 0 {
		return db
	}

	expr := "(" + strings.Join(clauses, " OR ") + ")"
	if negate {
		return db.Not(expr, args...)
	}
	return db.Where(expr, args...)
}

// buildYearClause returns a SQL fragment and args for a year value.
// Supports exact match ("1990") and range ("1990-2000").
func buildYearClause(s string) (clause string, args []interface{}, ok bool) {
	if lo, hi, found := strings.Cut(s, "-"); found {
		loY, err1 := strconv.Atoi(lo)
		hiY, err2 := strconv.Atoi(hi)
		if err1 != nil || err2 != nil {
			return
		}
		if loY > hiY {
			loY, hiY = hiY, loY
		}
		return "year BETWEEN ? AND ?", []interface{}{loY, hiY}, true
	}
	y, err := strconv.Atoi(s)
	if err != nil {
		return
	}
	return "year = ?", []interface{}{y}, true
}
