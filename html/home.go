package html

import (
	. "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"

	"maragu.dev/goo/html"

	"maragu.dev/goat/model"
)

func HomePage(props html.PageProps, cs []model.Conversation) Node {
	props.Title = "goat"

	return Page(props,
		Ol(
			Map(cs, func(c model.Conversation) Node {
				linkText := c.ID.String()
				if c.Topic != "" {
					linkText = c.Topic
				}
				return Li(A(Href("/conversations?id="+c.ID.String()), Text(linkText)))
			}),
		),
	)
}
