package html

import (
	_ "embed"

	g "github.com/maragudk/gomponents"
	c "github.com/maragudk/gomponents/components"
	. "github.com/maragudk/gomponents/html"

	"maragu.dev/goo/html"
)

//go:embed tailwind.config.js
var tailwindConfig string

func Page(props html.PageProps, children ...g.Node) g.Node {
	return c.HTML5(c.HTML5Props{
		Title:       props.Title,
		Description: props.Description,
		Language:    "en",
		Head: []g.Node{
			Script(Src("https://cdn.tailwindcss.com?plugins=forms,typography")),
			Script(g.Raw(tailwindConfig)),
		},
		Body: []g.Node{Class("bg-secondary font-serif"),
			Div(Class("min-h-screen flex flex-col justify-between bg-white"),
				header(),
				Div(Class("grow"),
					container(true, true,
						Div(Class("prose"),
							g.Group(children),
						),
					),
				),
				footer(),
			),
		},
	})
}

func header() g.Node {
	return Div(Class("bg-secondary shadow text-white"),
		container(true, false,
			Div(Class("h-14 flex items-center justify-between"),
				Img(Src("/images/logo.jpg"), Alt("Logo"), Class("h-12 w-auto")),
			),
		),
	)
}

func container(padX, padY bool, children ...g.Node) g.Node {
	return Div(
		c.Classes{
			"max-w-7xl mx-auto":     true,
			"px-4 md:px-8 lg:px-16": padX,
			"py-4 md:py-8":          padY,
		},
		g.Group(children),
	)
}

func footer() g.Node {
	return Div(Class("bg-secondary h-8"))
}
