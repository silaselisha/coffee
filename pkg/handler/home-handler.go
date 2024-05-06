package handler

import (
	"context"
	"net/http"
)

func (tmpl *Templates) RenderHomePageHandler(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	vars := struct {
		Name string
	}{Name: "HOME PAGE"}
	return render(tmpl.templates, w, "main", vars)

}
