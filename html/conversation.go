package html

import (
	"strings"

	"github.com/yuin/goldmark"
	. "maragu.dev/gomponents"
	hx "maragu.dev/gomponents-htmx"
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

		Group{
			H1(Text(title)),

			Div(Class("space-y-8"), hx.Get("/conversations?id="+cd.Conversation.ID.String()), hx.Trigger("every 1s"),
				TurnsPartial(cd),
			),

			P(A(Href("/"), Text("Back"))),
		},
	)
}

func TurnsPartial(cd model.ConversationDocument) Node {
	return Map(cd.Turns, func(t model.Turn) Node {
		s := cd.Speakers[t.SpeakerID]

		var content string
		var b strings.Builder
		if err := goldmark.Convert([]byte(t.Content), &b); err != nil {
			content = "Error converting markdown to HTML: " + err.Error()
		} else {
			content = b.String()
		}

		var reverse string
		if s.ID != model.MySpeakerID {
			reverse = " flex-row-reverse"
		}

		return Div(Class("flex"+reverse),
			P(Title(s.Name), Text(s.Avatar())),
			Div(Class("border border-gray-200 rounded-lg w-full px-4 mx-4"), Raw(content)),
		)
	})
}
