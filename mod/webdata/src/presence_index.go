package webdata

import (
	"cmp"
	"github.com/gin-gonic/gin"
	"net/http"
	"slices"
)

type presenceInfo struct {
	DisplayName string
	Key         string
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
			Key:         p.Identity.PublicKeyHex(),
			Alias:       p.Alias,
			Flags:       p.Flags,
		})
	}

	slices.SortFunc(page.Presence, func(a, b presenceInfo) int {
		return cmp.Compare(a.DisplayName, b.DisplayName)
	})

	c.HTML(http.StatusOK, "presence.index.gohtml", &page)
}
