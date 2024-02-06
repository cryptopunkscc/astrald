package webdata

import (
	"fmt"
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/data"
	"github.com/cryptopunkscc/astrald/log"
	"github.com/cryptopunkscc/astrald/mod/content"
	"github.com/cryptopunkscc/astrald/mod/fs"
	"github.com/cryptopunkscc/astrald/mod/keys"
	"github.com/cryptopunkscc/astrald/mod/media"
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
			DisplayName: m.DataID.String(),
			Type:        "unknown",
		}

		descs := mod.content.Describe(c, m.DataID, &content.DescribeOpts{
			IdentityFilter: id.AllowEveryone,
		})
		best := slicesSelect(descs, mod.preferredDesc)
		if best != nil {
			if s, ok := best.Data.(fmt.Stringer); ok {
				obj.DisplayName = s.String()
			}
		}

		for _, d := range descs {
			if td, ok := d.Data.(content.TypeDescriptor); ok {
				obj.Type = td.Type
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

func slicesSelect[T any](list []T, pick func(a, b T) T) T {
	var zero T
	if len(list) == 0 {
		return zero
	}
	if len(list) == 1 {
		return list[0]
	}
	best := list[0]
	for _, next := range list[1:] {
		best = pick(best, next)
	}
	return best
}

func (mod *Module) preferredDesc(a, b *content.Descriptor) *content.Descriptor {
	as := mod.scoreIdentity(a.Source) + mod.scoreType(a.Data.DescriptorType())
	bs := mod.scoreIdentity(b.Source) + mod.scoreType(b.Data.DescriptorType())
	if bs > as {
		return b
	}
	return a

}

func (mod *Module) scoreType(t string) int {
	var m = map[string]int{
		content.LabelDescriptor{}.DescriptorType(): 100,
		media.Descriptor{}.DescriptorType():        90,
		keys.KeyDescriptor{}.DescriptorType():      90,
		fs.FileDescriptor{}.DescriptorType():       80,
		content.TypeDescriptor{}.DescriptorType():  -100,
	}

	return m[t]
}

func (mod *Module) scoreIdentity(identity id.Identity) int {
	if identity.IsEqual(mod.node.Identity()) {
		return 9
	}

	return 0
}
