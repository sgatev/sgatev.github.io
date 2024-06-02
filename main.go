package main

import (
	"fmt"
	"github.com/alecthomas/chroma"
	chromahtml "github.com/alecthomas/chroma/formatters/html"
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
	"text/template"
)

var (
	htmlCodeFormatter *chromahtml.Formatter
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
	indexTempl, err := template.New("index.html").ParseFiles("templates/index.html")
	if err != nil {
		log.Fatal(err)
	}

	postTempl, err := template.New("post.html").ParseFiles("templates/post.html")
	if err != nil {
		log.Fatal(err)
	}

	m := minify.New()
	m.AddFunc("text/css", css.Minify)
	m.AddFunc("text/html", minifyhtml.Minify)

	processIndex := func(out string, args map[string]string) (err error) {
		w, err := os.Create(out)
		if err != nil {
			return err
		}
		defer func() {
			err = w.Close()
		}()

		mw := m.Writer("text/html", w)
		defer func() {
			err = mw.Close()
		}()

		return indexTempl.Execute(mw, args)
	}

	processPost := func(in, out string, args map[string]string) (err error) {
		md, err := ioutil.ReadFile(in)
		if err != nil {
			return err
		}

		args["Content"] = string(mdToHtml(md))

		w, err := os.Create(out)
		if err != nil {
			return err
		}
		defer func() {
			err = w.Close()
		}()

		mw := m.Writer("text/html", w)
		defer func() {
			err = mw.Close()
		}()

		return postTempl.Execute(mw, args)
	}

	htmlCodeFormatter = chromahtml.New(
		chromahtml.WithClasses(true),
		chromahtml.TabWidth(2))
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
		in := filepath.Join(postsDir, post.Name())
		out := filepath.Join(genDir, strings.Replace(post.Name(), "md", "html", 1))
		args := map[string]string{
			"Style": codeHighlightStyles.String(),
		}
		if err := processPost(in, out, args); err != nil {
			log.Fatal(err)
		}
	}

	if err := processIndex(filepath.Join(genDir, "index.html"), map[string]string{}); err != nil {
		log.Fatal(err)
	}
}
