package webdata

import (
	"encoding/json"
	"fmt"
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/data"
	"github.com/cryptopunkscc/astrald/log"
	"github.com/cryptopunkscc/astrald/mod/content"
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
)

type desc struct {
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
	Sets        []string
	Descs       []desc
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

	descs := mod.content.Describe(c, dataID, &content.DescribeOpts{
		IdentityFilter: id.AllowEveryone,
	})

	where, _ := mod.sets.Where(dataID)

	var page = objectsShowPage{
		DataID:      dataID,
		DisplayName: dataID.String(),
		Sets:        where,
	}

	best := slicesSelect(descs, mod.preferredDesc)
	if best != nil {
		if s, ok := best.Data.(fmt.Stringer); ok {
			page.DisplayName = s.String()
		}
	}

	for i, d := range descs {
		if td, ok := d.Data.(content.TypeDescriptor); ok {
			page.Type = td.Type
		}

		sourceName := mod.node.Resolver().DisplayName(d.Source)

		j, err := json.MarshalIndent(d.Data, "", "  ")
		if err != nil {
			c.Status(http.StatusInternalServerError)
			return
		}

		page.Descs = append(page.Descs, desc{
			ID:     "desc_" + strconv.FormatInt(int64(i), 10),
			Source: sourceName,
			Type:   d.Data.DescriptorType(),
			Data:   d.Data,
			JSON:   string(j),
		})
	}

	page.DisplayName = mod.parseIdentities(page.DisplayName)

	c.HTML(http.StatusOK, "objects.show.gohtml", &page)
}
