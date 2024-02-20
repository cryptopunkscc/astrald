package webdata

import (
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/data"
	"github.com/cryptopunkscc/astrald/lib/desc"
	"github.com/cryptopunkscc/astrald/log"
	"github.com/cryptopunkscc/astrald/mod/content"
	"github.com/cryptopunkscc/astrald/mod/sets"
	"github.com/cryptopunkscc/astrald/node"
	"github.com/gin-gonic/gin"
	"net/http"
)

type object struct {
	DataID      data.ID
	DisplayName string
	Type        string
}

func (o *object) SizeHuman() string {
	return log.DataSize(o.DataID.Size).HumanReadable()
}

type setShort struct {
	Name        string
	DisplayName string
}

type setPage struct {
	Name        string
	DisplayName string
	Count       int
	TotalSize   uint64
	Type        string
	SubsetCount int
	Subsets     []setShort
	Objects     []*object
}

func (p *setPage) TotalSizeHuman() string {
	return log.DataSize(p.TotalSize).HumanReadable()
}

func (mod *Module) handleSetsShow(c *gin.Context) {
	setName := c.Param("name")

	set, err := mod.sets.Open(setName, false)
	if err != nil {
		c.Status(http.StatusInternalServerError)
		return
	}

	stat, err := set.Stat()
	if err != nil {
		c.Status(http.StatusInternalServerError)
		return
	}

	members, err := set.Scan(nil)
	if err != nil {
		c.Status(http.StatusInternalServerError)
		return
	}

	var page = setPage{
		DisplayName: node.FormatString(mod.node, set.DisplayName()),
		Name:        setName,
		Count:       stat.Size,
		TotalSize:   stat.DataSize,
		Type:        string(stat.Type),
	}

	for _, m := range members {
		obj := &object{
			DataID:      m.DataID,
			DisplayName: mod.content.BestTitle(m.DataID),
			Type:        "unknown",
		}

		descs := mod.content.Describe(c, m.DataID, &desc.Opts{
			IdentityFilter: id.AllowEveryone,
		})

		// find the type descriptor
		for _, d := range descs {
			if td, ok := d.Data.(content.TypeDesc); ok {
				obj.Type = td.ContentType
			}
		}

		obj.DisplayName = node.FormatString(mod.node, obj.DisplayName)
		page.Objects = append(page.Objects, obj)
	}

	if stat.Type == sets.TypeUnion {
		union, ok := set.(sets.Union)
		if !ok {
			c.Status(http.StatusInternalServerError)
			return
		}

		subsets, _ := union.Subsets()
		for _, subname := range subsets {
			sub, err := mod.sets.Open(subname, false)
			if err != nil {
				page.Subsets = append(page.Subsets, setShort{
					Name:        subname,
					DisplayName: subname,
				})
				continue
			}
			page.Subsets = append(
				page.Subsets,
				setShort{
					Name:        subname,
					DisplayName: node.FormatString(mod.node, sub.DisplayName()),
				},
			)
		}
		page.SubsetCount = len(page.Subsets)
	}

	c.HTML(http.StatusOK, "sets.show.gohtml", &page)
}
