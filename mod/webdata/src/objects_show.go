package webdata

import (
	"cmp"
	"encoding/json"
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/data"
	"github.com/cryptopunkscc/astrald/lib/desc"
	"github.com/cryptopunkscc/astrald/log"
	"github.com/cryptopunkscc/astrald/mod/content"
	"github.com/cryptopunkscc/astrald/node"
	"github.com/gin-gonic/gin"
	"net/http"
	"slices"
	"strconv"
)

type descItem struct {
	ID     string
	Source string
	Type   string
	JSON   string
	Data   any
}

type objectsShowPage struct {
	DisplayName string
	DataID      data.ID
	Type        string
	Sets        []setShort
	Descs       []descItem
}

func (p objectsShowPage) SizeHuman() string {
	return log.DataSize(p.DataID.Size).HumanReadable()
}

func (mod *Module) handleObjectsShow(c *gin.Context) {
	dataID, err := data.Parse(c.Param("id"))
	if err != nil {
		c.Status(http.StatusBadRequest)
		return
	}

	descs := mod.content.Describe(c, dataID, &desc.Opts{
		IdentityFilter: id.AllowEveryone,
	})

	where, _ := mod.sets.Where(dataID)

	var page = objectsShowPage{
		DataID:      dataID,
		DisplayName: mod.content.BestTitle(dataID),
	}

	for _, setName := range where {
		set, err := mod.sets.Open(setName, false)
		if err != nil {
			continue
		}
		page.Sets = append(page.Sets, setShort{
			Name:        setName,
			DisplayName: node.FormatString(mod.node, set.DisplayName()),
		})

	}

	slices.SortFunc(page.Sets, func(a, b setShort) int {
		return cmp.Compare(a.DisplayName, b.DisplayName)
	})

	for i, d := range descs {
		if td, ok := d.Data.(content.TypeDesc); ok {
			page.Type = td.ContentType
		}

		sourceName := mod.node.Resolver().DisplayName(d.Source)

		j, err := json.MarshalIndent(d.Data, "", "  ")
		if err != nil {
			c.Status(http.StatusInternalServerError)
			return
		}

		page.Descs = append(page.Descs, descItem{
			ID:     "desc_" + strconv.FormatInt(int64(i), 10),
			Source: sourceName,
			Type:   d.Data.Type(),
			Data:   d.Data,
			JSON:   string(j),
		})
	}

	page.DisplayName = node.FormatString(mod.node, page.DisplayName)

	c.HTML(http.StatusOK, "objects.show.gohtml", &page)
}
