package webdata

import (
	"context"
	"github.com/cryptopunkscc/astrald/data"
	data2 "github.com/cryptopunkscc/astrald/mod/data"
	"html/template"
	"net/http"
	"path"
	"time"
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

	data, err := res.ReadFile("res/index.gohtml")
	if err != nil {
		panic(err)
	}

	handler.template, err = template.New("index").Parse(string(data))
	if err != nil {
		panic(err)
	}

	return handler
}

func (mod *IndexHandler) handleRequest(w http.ResponseWriter, r *http.Request) {
	var indexName = path.Base(r.URL.Path)

	list, err := mod.index.UpdatedBetween(indexName, time.Time{}, time.Time{})
	switch err {
	case nil:
	default:
		mod.log.Errorv(1, "error reading index: %v", err)
		w.WriteHeader(http.StatusNotFound)
		return
	}

	var page = &IndexPage{
		IndexName: indexName,
	}

	for _, item := range list {
		if !item.Added {
			continue
		}

		var entry = Entry{
			DataID: item.DataID,
			Label:  item.DataID.String(),
		}

		descs := mod.data.DescribeData(context.Background(), item.DataID, nil)
		for _, desc := range descs {
			switch typed := desc.Data.(type) {
			case data2.LabelDescriptor:
				entry.Label = typed.Label
			case data2.TypeDescriptor:
				entry.Type = typed.Type
			}
		}

		page.Entries = append(page.Entries, entry)
	}

	err = mod.template.Execute(w, page)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}

	return
}
