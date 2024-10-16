package http

import (
	"embed"

	"maragu.dev/httph"
	"maragu.dev/snorkel"

	"maragu.dev/goat/sql"
	goohttp "maragu.dev/goo/http"
)

func InjectHTTPRouter(log *snorkel.Logger, db *sql.Database, public embed.FS) func(*goohttp.Router) {
	return func(r *goohttp.Router) {
		// Group for HTML
		r.Group(func(r *goohttp.Router) {
			r.Use(httph.ContentSecurityPolicy(func(opts *httph.ContentSecurityPolicyOptions) {
				opts.ScriptSrc = "'self' 'unsafe-inline' https://cdn.tailwindcss.com"
				opts.StyleSrc = "'self' 'unsafe-inline'"
			}))

			Home(r, log, db)
			Conversation(r, log, db)
			Embedded(r, public)
		})
	}
}
