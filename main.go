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

func mdToHtml(md string) []byte {
	p := parser.NewWithExtensions(
		parser.CommonExtensions |
			parser.AutoHeadingIDs |
			parser.NoEmptyLineBeforeBlock)
	doc := p.Parse([]byte(md))
	r := html.NewRenderer(html.RendererOptions{
		Flags:          html.CommonFlags | html.HrefTargetBlank,
		RenderNodeHook: mdToHtmlRenderHook,
	})
	return markdown.Render(doc, r)
}

func main() {
	if err := os.Mkdir("gen", os.ModePerm); err != nil {
		log.Fatal(err)
	}
	if err := ioutil.WriteFile("gen/index.html", []byte(indexHtmlContent), 0644); err != nil {
		log.Fatal(err)
	}

	postHtmlContent := fmt.Sprintf(postHtmlTemplate, mdToHtml(postMdContent))
	if err := ioutil.WriteFile("gen/post.html", []byte(postHtmlContent), 0644); err != nil {
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

const postMdContent = `# Lorem Ipsum

## Lorem ipsum dolor sit

Lorem ipsum dolor sit amet, consectetur adipiscing elit. Integer
vestibulum commodo ligula vitae laoreet. Nam mollis in nisi eget
vulputate. Nulla eget euismod diam, eu tincidunt velit. Vivamus
tincidunt odio lobortis elit laoreet, a imperdiet tellus congue. In id
neque ut mauris suscipit condimentum. Phasellus pretium lacinia dapibus.
Vivamus at nisi non nulla posuere malesuada at non justo. Suspendisse in
maximus quam, id mollis elit. Vestibulum tempor rhoncus ante, ut
sollicitudin lacus gravida id. Nullam ac orci rutrum, vulputate velit
at, maximus nunc. Vestibulum mollis lectus eget mauris placerat, at
gravida tellus ullamcorper. Praesent suscipit nulla et auctor accumsan.
Suspendisse ultricies convallis est, sed tristique dui egestas ut.
Maecenas convallis scelerisque orci vel mollis. In condimentum, lorem
quis aliquet facilisis, dui dolor efficitur metus, eget elementum turpis
lectus non urna. Curabitur porta felis id dignissim pulvinar.

## Nulla euismod pellentesque

Nulla euismod pellentesque vehicula. Nulla purus elit, mattis id rhoncus
eu, fermentum efficitur nisl. Nunc mollis lorem non leo accumsan
pretium. Fusce in tempus dui. Interdum et malesuada fames ac ante ipsum
primis in faucibus. Aliquam vel libero vitae augue luctus hendrerit.
Nulla tincidunt facilisis lorem, sed congue mauris tempus ac. Vivamus
egestas ante id leo imperdiet gravida sit amet tristique dolor. Proin
feugiat magna eu tortor convallis scelerisque in in dolor. Cras risus
sapien, posuere quis nisi nec, dapibus laoreet lorem. Integer ut felis
non urna tincidunt dapibus ut nec purus. Duis justo diam, hendrerit
vitae nibh ut, eleifend molestie lacus. Sed at eros non ipsum
scelerisque lacinia vel sed est. Quisque et metus fermentum, dignissim
arcu non, placerat diam.

## Morbi ac est accumsan

Morbi ac est accumsan, mattis justo vel, sodales est. Morbi interdum,
ipsum ac malesuada accumsan, odio turpis sagittis purus, nec eleifend
arcu purus et arcu. Nam tristique fringilla lectus, vitae molestie eros
ultricies vel. Curabitur tincidunt erat vel eros tincidunt egestas.
Integer aliquet condimentum elit. Maecenas viverra dolor vehicula,
consectetur libero a, mollis ante. Nam nec feugiat enim, id congue ex.
Nulla metus enim, iaculis eget nunc sit amet, blandit congue justo.
Proin nec neque eu metus tincidunt porta. Nulla hendrerit urna semper
quam vehicula, eu hendrerit sapien accumsan. Nullam venenatis tortor et
semper semper. Maecenas sit amet mollis elit. Nullam in congue elit.
Integer tincidunt nec justo non posuere. Donec vel erat in turpis
tincidunt euismod.
`
