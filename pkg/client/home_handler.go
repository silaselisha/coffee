package client

import (
	"context"
	"net/http"
)

func (tmpl *Templates) RenderHomePageHandler(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	vars := struct {
		Name string
	}{Name: "HOME PAGE"}
	return wrietWebPage(tmpl.templates, w, "home", vars)
}
