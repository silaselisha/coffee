package handler

import (
	"context"
	"net/http"
	"text/template"
)

type Querier interface {
	RenderHomePageHandler(ctx context.Context, w http.ResponseWriter, r *http.Request) error
}

type Templates struct {
	templates *template.Template
}

func NewTemplate() Querier {
	return &Templates{
		templates: template.Must(template.ParseGlob("views/**/*.html")),
	}
}

func wrietWebPage(tmpl *template.Template, w http.ResponseWriter, name string, vars interface{}) error {
	// set cookies && sessions
	err := tmpl.ExecuteTemplate(w, name, vars)
	if err != nil {
		http.Error(w, "Failed to load template: "+err.Error(), http.StatusInternalServerError)
		return err
	}
	return nil
}
