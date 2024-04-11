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

	userID := mod.user.UserID()
	if userID.IsZero() {
		c.Status(http.StatusInternalServerError)
		return
	}

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
