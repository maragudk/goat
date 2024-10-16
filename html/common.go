package html

import (
	_ "embed"

	. "maragu.dev/gomponents"
	. "maragu.dev/gomponents/components"
	. "maragu.dev/gomponents/html"

	"maragu.dev/goo/html"
)

//go:embed tailwind.config.js
var tailwindConfig string

func Page(props html.PageProps, children ...Node) Node {
	return HTML5(HTML5Props{
		Title:       props.Title,
		Description: props.Description,
		Language:    "en",
		Head: []Node{
			Script(Src("https://cdn.tailwindcss.com?plugins=forms,typography")),
			Script(Src("https://unpkg.com/htmx.org@2/dist/htmx.min.js")),
			Script(Raw(tailwindConfig)),
		},
		Body: []Node{Class("bg-secondary font-serif"),
			Div(Class("min-h-screen flex flex-col justify-between bg-white"),
				header(),
				Div(Class("grow"),
					container(true, true,
						Div(Class("prose prose-lg"),
							Group(children),
						),
					),
				),
				footer(),
			),
		},
	})
}

func header() Node {
	return Div(Class("bg-secondary shadow text-white"),
		container(true, false,
			Div(Class("h-14 flex items-center justify-between"),
				A(Href("/"), Img(Src("/embedded/images/logo.jpg"), Alt("Logo"), Class("h-12 w-auto"))),
			),
		),
	)
}

func container(padX, padY bool, children ...Node) Node {
	return Div(
		Classes{
			"max-w-7xl mx-auto":     true,
			"px-4 md:px-8 lg:px-16": padX,
			"py-4 md:py-8":          padY,
		},
		Group(children),
	)
}

func footer() Node {
	return Div(Class("bg-secondary h-8"))
}
