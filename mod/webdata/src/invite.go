package webdata

import (
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/gin-gonic/gin"
	"net/http"
)

func (mod *Module) handleInvite(c *gin.Context) {
	nodeID, err := id.ParsePublicKeyHex(c.Param("id"))
	if err != nil {
		c.Status(http.StatusBadRequest)
		return
	}

	userIDs := mod.user.Identities()
	if len(userIDs) == 0 {
		c.Status(http.StatusInternalServerError)
		return
	}

	userID := userIDs[0]

	err = mod.setup.Invite(c, userID, nodeID)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"status": "error",
			"error":  err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "ok",
	})
}
