package webdata

import (
	"github.com/cryptopunkscc/astrald/log"
	"github.com/gin-gonic/gin"
	"net/http"
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
		stat, err := mod.sets.Stat(setName)
		if err != nil {
			c.Status(http.StatusInternalServerError)
			mod.log.Errorv(1, "sets.Stat %s: %v", setName, err)
			return
		}

		if !stat.Visible && !showHidden {
			continue
		}

		name := stat.Name
		if stat.Description != "" {
			name = stat.Description
		}

		list = append(list, setInfo{
			ID:       stat.Name,
			Name:     name,
			Size:     stat.Size,
			DataSize: log.DataSize(stat.DataSize).HumanReadable(),
			Type:     string(stat.Type),
		})
	}

	c.HTML(http.StatusOK, "sets.index.gohtml", gin.H{
		"sets": list,
	})
}
