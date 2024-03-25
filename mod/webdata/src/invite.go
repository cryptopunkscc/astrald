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

	user := mod.user.LocalUser()
	if user == nil {
		c.Status(http.StatusInternalServerError)
		return
	}

	err = mod.setup.Invite(c, user.Identity(), nodeID)
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
