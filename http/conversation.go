package http

import (
	"context"

	g "github.com/maragudk/gomponents"
	"maragu.dev/snorkel"

	"maragu.dev/goat/html"
	"maragu.dev/goat/model"
	goohtml "maragu.dev/goo/html"
	"maragu.dev/goo/http"
)

type conversationGetter interface {
	GetConversationDocument(ctx context.Context, id model.ID) (model.ConversationDocument, error)
}

func Conversation(r *http.Router, log *snorkel.Logger, db conversationGetter) {
	r.Get("/conversations", func(props goohtml.PageProps) (g.Node, error) {
		id := model.ID(props.Req.URL.Query().Get("id"))

		cd, err := db.GetConversationDocument(props.Ctx, id)
		if err != nil {
			log.Event("Error getting conversation document", 1, "error", err)
			return goohtml.ErrorPage(html.Page), err
		}

		return html.ConversationPage(props, cd), nil
	})
}
