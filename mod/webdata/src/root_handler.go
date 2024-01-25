package webdata

import (
	"html/template"
	"net/http"
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

	handler.template, err = template.New("index").Parse(string(data))
	if err != nil {
		panic(err)
	}

	return handler
}

func (mod *RootHandler) handleRequest(w http.ResponseWriter, r *http.Request) {
	list, err := mod.index.AllIndexes()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if len(list) > 0 {
		err = mod.template.Execute(w, list)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
		}
	}

	return
}
