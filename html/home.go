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
				linkText := cd.Conversation.ID.String()
				if cd.Conversation.Topic != "" {
					linkText = cd.Conversation.Topic
				}
				return Li(A(Href("/conversations?id="+cd.Conversation.ID.String()), Text(linkText)))
			}),
		),
	)
}
