package webdata

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

type presenceInfo struct {
	DisplayName string
	Alias       string
	Flags       []string
}

type presenceIndexPage struct {
	Title    string
	Presence []presenceInfo
}

func (mod *Module) handlePresenceIndex(c *gin.Context) {
	var page = presenceIndexPage{
		Title: "Presence",
	}

	for _, p := range mod.presence.List() {
		page.Presence = append(page.Presence, presenceInfo{
			DisplayName: mod.node.Resolver().DisplayName(p.Identity),
			Alias:       p.Alias,
			Flags:       p.Flags,
		})
	}

	c.HTML(http.StatusOK, "presence.index.gohtml", &page)
}
