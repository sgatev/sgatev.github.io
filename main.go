package main

import (
	"fmt"
	"github.com/alecthomas/chroma"
	"github.com/alecthomas/chroma/formatters/html"
	"github.com/alecthomas/chroma/lexers"
	"github.com/alecthomas/chroma/styles"
	"github.com/gomarkdown/markdown"
	"github.com/gomarkdown/markdown/ast"
	mdhtml "github.com/gomarkdown/markdown/html"
	"github.com/gomarkdown/markdown/parser"
	"github.com/tdewolff/minify/v2"
	"github.com/tdewolff/minify/v2/css"
	minifyhtml "github.com/tdewolff/minify/v2/html"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
)

var (
	htmlCodeFormatter *html.Formatter
	highlightStyle    *chroma.Style
)

func renderTitle(w io.Writer, entering bool) {
	if entering {
		io.WriteString(w, "<h1>")
	} else {
		io.WriteString(w, "</h1>")
		io.WriteString(w, "<hr />")
	}
}

func htmlCodeHighlight(w io.Writer, source, lang string) error {
	l := lexers.Get(lang)
	if l == nil {
		l = lexers.Analyse(source)
	}
	if l == nil {
		l = lexers.Fallback
	}
	l = chroma.Coalesce(l)

	it, err := l.Tokenise(nil, source)
	if err != nil {
		return err
	}
	return htmlCodeFormatter.Format(w, highlightStyle, it)
}

func renderCode(w io.Writer, codeBlock *ast.CodeBlock) {
	htmlCodeHighlight(w, string(codeBlock.Literal), string(codeBlock.Info))
}

func mdToHtmlRenderHook(
	w io.Writer, node ast.Node, entering bool) (ast.WalkStatus, bool) {

	if heading, ok := node.(*ast.Heading); ok {
		if heading.Level == 1 {
			renderTitle(w, entering)
			return ast.GoToNext, true
		}
	} else if code, ok := node.(*ast.CodeBlock); ok {
		renderCode(w, code)
		return ast.GoToNext, true
	}
	return ast.GoToNext, false
}

func mdToHtml(md []byte) []byte {
	p := parser.NewWithExtensions(
		parser.CommonExtensions |
			parser.AutoHeadingIDs |
			parser.NoEmptyLineBeforeBlock)
	doc := p.Parse(md)
	r := mdhtml.NewRenderer(mdhtml.RendererOptions{
		Flags:          mdhtml.CommonFlags | mdhtml.HrefTargetBlank,
		RenderNodeHook: mdToHtmlRenderHook,
	})
	return markdown.Render(doc, r)
}

func main() {
	m := minify.New()
	m.AddFunc("text/css", css.Minify)
	m.AddFunc("text/html", minifyhtml.Minify)

	htmlCodeFormatter = html.New(html.WithClasses(true), html.TabWidth(2))
	if htmlCodeFormatter == nil {
		log.Fatal("chroma: couldn't create HTML formatter")
	}

	const highlightStyleName = "bw"
	highlightStyle = styles.Get(highlightStyleName)
	if highlightStyle == nil {
		log.Fatal(fmt.Sprintf("chroma: couldn't find style '%s'", highlightStyleName))
	}

	var codeHighlightStyles strings.Builder
	htmlCodeFormatter.WriteCSS(&codeHighlightStyles, highlightStyle)

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

		postHtmlContent := fmt.Sprintf(postHtmlTemplate, codeHighlightStyles.String(), mdToHtml(postMdContent))
		postHtmlContent, err = m.String("text/html", postHtmlContent)
		if err != nil {
			log.Fatal(err)
		}

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

      footer {
        margin-bottom: 3rem;
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

      footer {
        margin-bottom: 3rem;
      }

      a {
        text-decoration: none;
        color: inherit;
      }
      %s
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
