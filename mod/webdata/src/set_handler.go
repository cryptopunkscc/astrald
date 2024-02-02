package webdata

import (
	"cmp"
	"context"
	"github.com/cryptopunkscc/astrald/data"
	"github.com/cryptopunkscc/astrald/mod/content"
	"github.com/cryptopunkscc/astrald/mod/fs"
	"github.com/cryptopunkscc/astrald/mod/keys"
	"github.com/cryptopunkscc/astrald/mod/media"
	"github.com/cryptopunkscc/astrald/mod/relay"
	"github.com/cryptopunkscc/astrald/mod/zip"
	"html/template"
	"net/http"
	"path"
	"path/filepath"
	"slices"
)

type SetHandler struct {
	*Module
	template *template.Template
}

type SetPage struct {
	SetName string
	Entries []*Entry
}

type Entry struct {
	DataID data.ID
	Label  string
	Type   string
}

func NewSetHandler(module *Module) *SetHandler {
	handler := &SetHandler{Module: module}

	bytes, err := res.ReadFile("res/set.gohtml")
	if err != nil {
		panic(err)
	}

	handler.template, err = template.New("set").Parse(string(bytes))
	if err != nil {
		panic(err)
	}

	return handler
}

func (mod *SetHandler) handleRequest(w http.ResponseWriter, r *http.Request) {
	var setName = path.Base(r.URL.Path)

	set, err := mod.sets.Open(setName)
	if err != nil {
		mod.log.Errorv(1, "error opening set: %v", err)
		w.WriteHeader(http.StatusNotFound)
		return
	}

	list, err := set.Scan(nil)
	if err != nil {
		mod.log.Errorv(1, "error scanning set: %v", err)
		w.WriteHeader(http.StatusNotFound)
		return
	}

	info, err := mod.sets.SetInfo(setName)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	var page = &SetPage{
		SetName: setName,
	}
	if info.Description != "" {
		page.SetName = info.Description
	}

	for _, item := range list {
		if item.Removed {
			continue
		}

		var entry = &Entry{
			DataID: item.DataID,
			Label:  item.DataID.String(),
		}

		descs := mod.content.Describe(context.Background(), item.DataID, nil)
		for _, desc := range descs {
			switch typed := desc.(type) {
			case content.LabelDescriptor:
				entry.Label = typed.Label
			case content.TypeDescriptor:
				entry.Type = typed.Type
			case media.Descriptor:
				entry.Label = typed.Artist + " - " + typed.Title
			case keys.KeyDescriptor:
				entry.Label = "Private key of " + mod.node.Resolver().DisplayName(typed.PublicKey)
			case relay.CertDescriptor:
				relayName := mod.node.Resolver().DisplayName(typed.RelayID)
				targetName := mod.node.Resolver().DisplayName(typed.TargetID)
				entry.Label = "Relay certificate for " + targetName + "@" + relayName
			case fs.FileDescriptor:
				if len(typed.Paths) == 0 {
					continue
				}
				entry.Label = filepath.Base(typed.Paths[0])
			case zip.MemberDescriptor:
				if len(typed.Memberships) == 0 {
					continue
				}
				entry.Label = filepath.Base(typed.Memberships[0].Path)
			}
		}

		page.Entries = append(page.Entries, entry)
	}

	slices.SortFunc(page.Entries, func(a, b *Entry) int {
		return cmp.Compare(a.Label, b.Label)
	})

	err = mod.template.Execute(w, page)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}

	return
}
