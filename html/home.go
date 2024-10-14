package html

import (
	. "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"

	"maragu.dev/goat/model"
	"maragu.dev/goo/html"
)

func HomePage(props html.PageProps, cds []model.ConversationDocument) Node {
	props.Title = "goat"

	return Page(props,
		Ol(
			Map(cds, func(cd model.ConversationDocument) Node {
				return Li(A(Href("/conversations?id="+cd.Conversation.ID.String()), Text(cd.Conversation.ID.String())))
			}),
		),
	)
}
