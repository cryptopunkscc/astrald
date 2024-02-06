package webdata

import (
	"cmp"
	"github.com/cryptopunkscc/astrald/log"
	"github.com/cryptopunkscc/astrald/node"
	"github.com/gin-gonic/gin"
	"net/http"
	"slices"
)

type setInfo struct {
	ID       string
	Name     string
	Size     int
	Type     string
	DataSize string
}

func (mod *Module) handleSetsIndex(c *gin.Context) {
	allSets, err := mod.sets.All()
	if err != nil {
		c.Status(http.StatusInternalServerError)
		return
	}

	var showHidden = len(c.Query("hidden")) > 0

	var list []setInfo
	for _, setName := range allSets {
		if setName[0] == '.' && !showHidden {
			continue
		}
		set, err := mod.sets.Open(setName, false)
		if err != nil {
			continue
		}

		stat, err := set.Stat()
		if err != nil {
			c.Status(http.StatusInternalServerError)
			mod.log.Errorv(1, "sets.Stat %s: %v", setName, err)
			return
		}

		list = append(list, setInfo{
			ID:       stat.Name,
			Name:     node.FormatString(mod.node, set.DisplayName()),
			Size:     stat.Size,
			DataSize: log.DataSize(stat.DataSize).HumanReadable(),
			Type:     string(stat.Type),
		})
	}

	slices.SortFunc(list, func(a, b setInfo) int {
		return cmp.Compare(a.Name, b.Name)
	})

	c.HTML(http.StatusOK, "sets.index.gohtml", gin.H{
		"sets": list,
	})
}
