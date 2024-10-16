package html

import (
	"strings"

	"github.com/yuin/goldmark"
	. "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"

	"maragu.dev/goat/model"
	"maragu.dev/goo/html"
)

func ConversationPage(props html.PageProps, cd model.ConversationDocument) Node {
	title := cd.Conversation.Topic
	if title == "" {
		title = cd.Conversation.ID.String()
	}

	props.Title = title

	return Page(props,
		Div(
			H1(Text(title)),

			Div(Class("space-y-8"),
				Map(cd.Turns, func(t model.Turn) Node {
					s := cd.Speakers[t.SpeakerID]

					var content string
					var b strings.Builder
					if err := goldmark.Convert([]byte(t.Content), &b); err != nil {
						content = "Error converting markdown to HTML: " + err.Error()
					} else {
						content = b.String()
					}

					return Div(Class("flex space-x-4"),
						P(Title(s.Name), Text(s.Avatar())),
						Div(Class("border border-gray-50 shadow rounded w-full p-4"), Raw(content)),
					)
				}),
			),
		),
	)
}
