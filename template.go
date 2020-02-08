package main

import (
	html "html/template"
	"io"
	"log"

	"github.com/mmcdole/gofeed"
)

const (
	mainTemplate = `
<!DOCTYPE html>
<html lang="en">
	<head>
		<meta charset="utf-8" >
		<meta name="description" content="{{ .Description }}">
		<title>{{ .Title }}</title>
	</head>
	<body>
		<header>
			<p>
				Title: <strong>{{ .Title }}</strong><br>
				Description: <q>{{ .Description }}</q><br>
				Link: <a href="{{.FeedLink}}" target="_blank">{{ .FeedLink }}</a><br>
				Updated: <q>{{ .Updated }}</q> <br>
			</p>
			<p>Categories:
				<ul>
						{{ range .Categories }}
							<li>{{ . }}</li>
						{{ end }}
				</ul>
			</p>
		</header>
		{{ range $item := .Items }}
		<section>
			<h2>{{ $item.Title }}</h2>
			<h3>{{ $item.Description }}</h3>
			<p>
				{{ $item }}
			</p>
		</section>
		{{ end }}
	</body>
</html>`
)

func generateMainTemplate(writer io.Writer, entries *gofeed.Feed) error {
	var funcs = html.FuncMap{}
	tpl, err := html.New("main").Funcs(funcs).Parse(mainTemplate)
	if err != nil {
		return err
	}

	for _, item := range entries.Items {
		log.Printf("%+v\n", *item)

	}

	err = tpl.Execute(writer, *entries)
	return err
}
