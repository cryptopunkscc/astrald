package webdata

import (
	"cmp"
	"github.com/cryptopunkscc/astrald/mod/sets"
	"html/template"
	"net/http"
	"slices"
)

type RootHandler struct {
	*Module
	template *template.Template
}

func NewRootHandler(module *Module) *RootHandler {
	var err error
	var handler = &RootHandler{Module: module}

	data, err := res.ReadFile("res/root.gohtml")
	if err != nil {
		panic(err)
	}

	handler.template, err = template.New("root").Parse(string(data))
	if err != nil {
		panic(err)
	}

	return handler
}

func (mod *RootHandler) handleRequest(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	var showHidden = r.URL.Query().Has("hidden")

	list, err := mod.sets.All()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	var filtered []sets.Info
	for _, i := range list {
		if !i.Visible && !showHidden {
			continue
		}
		filtered = append(filtered, i)
	}

	slices.SortFunc(filtered, func(a, b sets.Info) int {
		return cmp.Compare(a.Name, b.Name)
	})

	if len(filtered) > 0 {
		err = mod.template.Execute(w, filtered)
		if err != nil {
			mod.log.Errorv(1, "error rendering root template: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
		}
	}

	return
}
