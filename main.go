package main

import (
	"fmt"
	"github.com/gomarkdown/markdown"
	"github.com/gomarkdown/markdown/ast"
	"github.com/gomarkdown/markdown/html"
	"github.com/gomarkdown/markdown/parser"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
)

func renderTitle(w io.Writer, node ast.Node, entering bool) {
	if entering {
		io.WriteString(w, "<h1>")
	} else {
		io.WriteString(w, "</h1>")
		io.WriteString(w, "<hr />")
	}
}

func mdToHtmlRenderHook(w io.Writer, node ast.Node, entering bool) (ast.WalkStatus, bool) {
	if h, ok := node.(*ast.Heading); ok {
		if h.Level == 1 {
			renderTitle(w, h, entering)
			return ast.GoToNext, true
		}
	}
	return ast.GoToNext, false
}

func mdToHtml(md []byte) []byte {
	p := parser.NewWithExtensions(
		parser.CommonExtensions |
			parser.AutoHeadingIDs |
			parser.NoEmptyLineBeforeBlock)
	doc := p.Parse(md)
	r := html.NewRenderer(html.RendererOptions{
		Flags:          html.CommonFlags | html.HrefTargetBlank,
		RenderNodeHook: mdToHtmlRenderHook,
	})
	return markdown.Render(doc, r)
}

func main() {
	const genDir = "gen"
	if err := os.Mkdir(genDir, os.ModePerm); err != nil {
		log.Fatal(err)
	}

	const postsDir = "posts"
	posts, err := os.ReadDir(postsDir)
	if err != nil {
		log.Fatal(err)
	}

	for _, post := range posts {
		postMdContent, err := ioutil.ReadFile(filepath.Join(postsDir, post.Name()))
		if err != nil {
			log.Fatal(err)
		}

		postHtmlContent := fmt.Sprintf(postHtmlTemplate, mdToHtml(postMdContent))
		if err := ioutil.WriteFile(filepath.Join(genDir, "post.html"), []byte(postHtmlContent), 0644); err != nil {
			log.Fatal(err)
		}
	}

	if err := ioutil.WriteFile(filepath.Join(genDir, "index.html"), []byte(indexHtmlContent), 0644); err != nil {
		log.Fatal(err)
	}
}

const indexHtmlContent = `<!doctype html>
<html lang="en-US">
  <head>
    <meta charset="utf-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <meta name="description" content="Posts by Stanislav Gatev.">
    <title>Posts</title>
    <style>
      body {
        font-family: "Arial", sans-serif;
        line-height: 1.5rem;
      }

      h1,
      h2 {
        font-family: "Garamond", serif;
      }

      h1 {
        font-size: 2.5rem;
      }

      article,
      footer {
        max-width: 600px;
        margin-left: auto;
        margin-right: auto;
      }

      article {
        margin-top: 3rem;
      }

      a {
        text-decoration: none;
        color: inherit;
      }
    </style>
  </head>
  <body>
    <article>
      <h1>Posts</h1>
      <hr />
      <section>
        <h2><a href="/post.html">Lorem Ipsum</a></h2>
      </section>
    </article>
    <footer>
      <hr />
      2024 © Stanislav Gatev
    </footer>
  </body>
</html>
`

const postHtmlTemplate = `<!doctype html>
<html lang="en-US">
  <head>
    <meta charset="utf-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <meta name="description" content="Lorem Ipsum.">
    <title>Lorem Ipsum</title>
    <style>
      body {
        font-family: "Arial", sans-serif;
        line-height: 1.5rem;
      }

      h1,
      h2 {
        font-family: "Garamond", serif;
      }

      h1 {
        font-size: 2.5rem;
      }

      article,
      footer {
        max-width: 600px;
        margin-left: auto;
        margin-right: auto;
      }

      article {
        margin-top: 3rem;
      }

      a {
        text-decoration: none;
        color: inherit;
      }
    </style>
  </head>
  <body>
    <article>
      %s
    </article>
    <footer>
      <hr />
      2024 © Stanislav Gatev
    </footer>
  </body>
</html>
`
