package webdata

import (
	"fmt"
	"github.com/cryptopunkscc/astrald/data"
	"github.com/cryptopunkscc/astrald/mod/storage"
	"github.com/gin-gonic/gin"
	"io"
	"net/http"
	"strconv"
)

func (mod *Module) handleObjectsOpen(c *gin.Context) {
	dataID, err := data.Parse(c.Param("id"))
	if err != nil {
		c.Status(http.StatusBadRequest)
		return
	}

	c.Header("Accept-Ranges", "bytes")

	reqRange := c.GetHeader("Range")

	var opts = &storage.OpenOpts{Virtual: true, Network: true}
	var length = int64(dataID.Size)

	ranges, err := ParseRange(reqRange, int64(dataID.Size))
	if len(ranges) > 0 {
		opts.Offset = uint64(ranges[0].Start)
		length = ranges[0].Length
	}

	if !mod.node.Auth().Authorize(mod.identity, storage.OpenAction, dataID) {
		c.Status(http.StatusForbidden)
		return
	}

	reader, err := mod.storage.Open(dataID, opts)
	if err != nil {
		mod.log.Errorv(2, "read %v: %v", dataID, err)
		c.Status(http.StatusNotFound)
		return
	}
	defer reader.Close()

	c.Header("Content-Length", strconv.FormatInt(length, 10))

	if d := c.Query("download"); d != "" {
		c.Header(
			"Content-Disposition",
			fmt.Sprintf("attachment; filename=\"%s\"", dataID.String()),
		)
	}

	if opts.Offset > 0 || uint64(length) != dataID.Size {
		resRange := fmt.Sprintf("bytes %d-%d/%d",
			opts.Offset,
			opts.Offset+uint64(length)-1,
			dataID.Size,
		)

		c.Header("Content-Range", resRange)
		c.Status(http.StatusPartialContent)
	}

	io.CopyN(c.Writer, reader, length)
}
