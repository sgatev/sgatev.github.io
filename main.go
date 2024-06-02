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
	"strconv"
	"strings"
	"text/template"
	"time"
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

type htmlRenderer struct {
	m *minify.M
}

func (p *htmlRenderer) renderHtml(
	path string, templ *template.Template, args map[string]string) (err error) {

	w, err := os.Create(path)
	if err != nil {
		return err
	}
	defer func() {
		err = w.Close()
	}()

	mw := p.m.Writer("text/html", w)
	defer func() {
		err = mw.Close()
	}()

	return templ.Execute(mw, args)
}

func makeHtmlRenderer() *htmlRenderer {
	p := &htmlRenderer{}

	p.m = minify.New()
	p.m.AddFunc("text/css", css.Minify)
	p.m.AddFunc("text/html", minifyhtml.Minify)

	return p
}

func main() {
	indexTempl, err := template.ParseFiles(
		"templates/index.html", "templates/footer.html", "templates/structure.css")
	if err != nil {
		log.Fatal(err)
	}

	postTempl, err := template.ParseFiles(
		"templates/post.html", "templates/footer.html", "templates/structure.css")
	if err != nil {
		log.Fatal(err)
	}

	r := makeHtmlRenderer()

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

	var codeHighlightStyle strings.Builder
	htmlCodeFormatter.WriteCSS(&codeHighlightStyle, highlightStyle)

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
		md, err := ioutil.ReadFile(in)
		if err != nil {
			log.Fatal(err)
		}

		out := filepath.Join(genDir, strings.Replace(post.Name(), "md", "html", 1))
		args := map[string]string{
			"CodeHighlightStyle": codeHighlightStyle.String(),
			"Content":            string(mdToHtml(md)),
			"CurrentYear":        strconv.Itoa(time.Now().Year()),
		}

		if err := r.renderHtml(out, postTempl, args); err != nil {
			log.Fatal(err)
		}
	}

	indexOut := filepath.Join(genDir, "index.html")
	indexArgs := map[string]string{
		"CurrentYear": strconv.Itoa(time.Now().Year()),
	}
	if err := r.renderHtml(indexOut, indexTempl, indexArgs); err != nil {
		log.Fatal(err)
	}
}
