package webdata

import (
	"fmt"
	"github.com/cryptopunkscc/astrald/data"
	"github.com/cryptopunkscc/astrald/mod/storage"
	"io"
	"net/http"
	"path"
	"strconv"
)

type DataHandler struct {
	*Module
}

func NewDataHandler(module *Module) *DataHandler {
	return &DataHandler{Module: module}
}

func (mod *DataHandler) handleRequest(w http.ResponseWriter, r *http.Request) {
	dataID, err := data.Parse(path.Base(r.URL.Path))
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	w.Header().Set("Accept-Ranges", "bytes")

	reqRange := r.Header.Get("Range")

	var opts = &storage.OpenOpts{Virtual: true, Network: true}
	var length = int64(dataID.Size)

	ranges, err := ParseRange(reqRange, int64(dataID.Size))
	if len(ranges) > 0 {
		opts.Offset = uint64(ranges[0].Start)
		length = ranges[0].Length
	}

	if err = mod.shares.Authorize(mod.identity, dataID); err != nil {
		mod.log.Errorv(1, "denied %v access to %v: %v", mod.identity, dataID, err)
		w.WriteHeader(http.StatusForbidden)
		return
	}

	reader, err := mod.storage.Open(dataID, opts)
	if err != nil {
		mod.log.Errorv(2, "read %v: %v", dataID, err)
		w.WriteHeader(http.StatusNotFound)
		return
	}
	defer reader.Close()

	w.Header().Set("Content-Length", strconv.FormatInt(length, 10))

	if opts.Offset > 0 || uint64(length) != dataID.Size {
		resRange := fmt.Sprintf("bytes %d-%d/%d",
			opts.Offset,
			opts.Offset+uint64(length)-1,
			dataID.Size,
		)

		w.Header().Set("Content-Range", resRange)
		w.WriteHeader(http.StatusPartialContent)
	}

	io.CopyN(w, reader, length)
}
