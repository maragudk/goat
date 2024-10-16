package service

import (
	"context"
	"embed"
	"io"

	"maragu.dev/snorkel"

	"maragu.dev/goat/html"
	"maragu.dev/goat/http"
	"maragu.dev/goat/sql"
	gooservice "maragu.dev/goo/service"
)

func (s *Service) Serve(ctx context.Context, db *sql.Database, public embed.FS, err io.Writer) {
	log := snorkel.New(snorkel.Options{W: err})
	gooservice.Start(gooservice.Options{
		HTMLPage:           html.Page,
		HTTPRouterInjector: http.InjectHTTPRouter(log, db, public),
		Log:                log,
		SQLHelperInjector:  db.InjectSQLHelper,
	})
}
