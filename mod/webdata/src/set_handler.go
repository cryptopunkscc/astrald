package webdata

import (
	"cmp"
	"context"
	"github.com/cryptopunkscc/astrald/data"
	"github.com/cryptopunkscc/astrald/mod/content"
	"github.com/cryptopunkscc/astrald/mod/media"
	"html/template"
	"net/http"
	"path"
	"slices"
)

type SetHandler struct {
	*Module
	template *template.Template
}

type SetPage struct {
	SetName string
	Entries []Entry
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

		var entry = Entry{
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
			}
		}

		page.Entries = append(page.Entries, entry)
	}

	slices.SortFunc(page.Entries, func(a, b Entry) int {
		return cmp.Compare(a.Label, b.Label)
	})

	err = mod.template.Execute(w, page)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}

	return
}
