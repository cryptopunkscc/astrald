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

type IndexHandler struct {
	*Module
	template *template.Template
}

type IndexPage struct {
	IndexName string
	Entries   []Entry
}

type Entry struct {
	DataID data.ID
	Label  string
	Type   string
}

func NewIndexHandler(module *Module) *IndexHandler {
	handler := &IndexHandler{Module: module}

	bytes, err := res.ReadFile("res/index.gohtml")
	if err != nil {
		panic(err)
	}

	handler.template, err = template.New("index").Parse(string(bytes))
	if err != nil {
		panic(err)
	}

	return handler
}

func (mod *IndexHandler) handleRequest(w http.ResponseWriter, r *http.Request) {
	var indexName = path.Base(r.URL.Path)

	list, err := mod.index.Scan(indexName, nil)
	switch err {
	case nil:
	default:
		mod.log.Errorv(1, "error reading index: %v", err)
		w.WriteHeader(http.StatusNotFound)
		return
	}

	info, err := mod.index.IndexInfo(indexName)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	var page = &IndexPage{
		IndexName: indexName,
	}
	if info.Description != "" {
		page.IndexName = info.Description
	}

	for _, item := range list {
		if !item.Added {
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
