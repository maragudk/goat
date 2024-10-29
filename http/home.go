package http

import (
	"context"

	. "maragu.dev/gomponents"
	"maragu.dev/snorkel"

	goohtml "maragu.dev/goo/html"
	goohttp "maragu.dev/goo/http"

	"maragu.dev/goat/html"
	"maragu.dev/goat/model"
)

type conversationsGetter interface {
	GetConversations(ctx context.Context) ([]model.Conversation, error)
}

func Home(r *goohttp.Router, log *snorkel.Logger, db conversationsGetter) {
	r.Get("/", func(props goohtml.PageProps) (Node, error) {
		cs, err := db.GetConversations(props.Ctx)
		if err != nil {
			log.Event("Error getting conversations", 1, "error", err)
			return goohtml.ErrorPage(html.Page), err
		}

		return html.HomePage(props, cs), nil
	})
}
