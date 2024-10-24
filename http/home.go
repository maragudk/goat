package http

import (
	"context"

	. "maragu.dev/gomponents"
	"maragu.dev/snorkel"

	"maragu.dev/goat/html"
	"maragu.dev/goat/model"
	goohtml "maragu.dev/goo/html"
	goohttp "maragu.dev/goo/http"
)

type conversationsGetter interface {
	GetConversationDocuments(ctx context.Context) ([]model.ConversationDocument, error)
}

func Home(r *goohttp.Router, log *snorkel.Logger, db conversationsGetter) {
	r.Get("/", func(props goohtml.PageProps) (Node, error) {
		cds, err := db.GetConversationDocuments(props.Ctx)
		if err != nil {
			log.Event("Error getting conversation documents", 1, "error", err)
			return goohtml.ErrorPage(html.Page), err
		}

		return html.HomePage(props, cds), nil
	})
}
