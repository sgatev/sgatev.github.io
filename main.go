package main

import (
	"bytes"
	"fmt"
	"github.com/alecthomas/chroma/v2"
	chromahtml "github.com/alecthomas/chroma/v2/formatters/html"
	"github.com/alecthomas/chroma/v2/lexers"
	"github.com/alecthomas/chroma/v2/styles"
	"github.com/gomarkdown/markdown"
	"github.com/gomarkdown/markdown/ast"
	mdhtml "github.com/gomarkdown/markdown/html"
	"github.com/gomarkdown/markdown/parser"
	"github.com/tdewolff/minify/v2"
	"github.com/tdewolff/minify/v2/css"
	minifyhtml "github.com/tdewolff/minify/v2/html"
	"github.com/tdewolff/minify/v2/js"
	"gopkg.in/yaml.v2"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"text/template"
	"time"
)

var (
	htmlCodeFormatter *chromahtml.Formatter
	highlightStyle    *chroma.Style
)

type metadata struct {
	Title string    `yaml:"title"`
	Date  time.Time `yaml:"date"`
}

func extractMetadata(content []byte) (*metadata, []byte, error) {
	parts := bytes.Split(content, []byte("---"))
	if len(parts) != 3 {
		return nil, nil, fmt.Errorf("Invalid syntax")
	}

	var meta metadata
	if err := yaml.Unmarshal(parts[1], &meta); err != nil {
		return nil, nil, err
	}

	return &meta, parts[2], nil
}

type post struct {
	Title string
	Path  string
	Date  time.Time
}

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

type htmlRenderer struct {
	m *minify.M
}

func (p *htmlRenderer) renderHtml(
	path string, templ *template.Template, args any) error {

	var s strings.Builder
	if err := templ.Execute(&s, args); err != nil {
		return err
	}

	ms, err := p.m.String("text/html", s.String())
	if err != nil {
		return err
	}

	return ioutil.WriteFile(path, []byte(ms), 0644)
}

func (p *htmlRenderer) renderCss(
	path string, templ *template.Template, args any) error {

	var s strings.Builder
	if err := templ.Execute(&s, args); err != nil {
		return err
	}

	ms, err := p.m.String("text/css", s.String())
	if err != nil {
		return err
	}

	return ioutil.WriteFile(path, []byte(ms), 0644)
}

func (p *htmlRenderer) renderJs(
	path string, templ *template.Template, args any) error {

	var s strings.Builder
	if err := templ.Execute(&s, args); err != nil {
		return err
	}

	ms, err := p.m.String("text/js", s.String())
	if err != nil {
		return err
	}

	return ioutil.WriteFile(path, []byte(ms), 0644)
}

func makeHtmlRenderer() *htmlRenderer {
	p := &htmlRenderer{}

	p.m = minify.New()
	p.m.AddFunc("text/css", css.Minify)
	p.m.Add("text/html", &minifyhtml.Minifier{
		KeepEndTags: true,
	})
	p.m.AddFunc("text/js", js.Minify)

	return p
}

func makePostTemplate(layoutTempl *template.Template) *template.Template {
	return layoutTempl.Lookup("layout.html")
}

func main() {
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

	const genDir = "gen"
	if err := os.Mkdir(genDir, os.ModePerm); err != nil {
		log.Fatal(err)
	}

	const postsDir = "posts"
	postFiles, err := os.ReadDir(postsDir)
	if err != nil {
		log.Fatal(err)
	}

	var posts []post
	for _, postFile := range postFiles {
		in := filepath.Join(postsDir, postFile.Name())
		md, err := ioutil.ReadFile(in)
		if err != nil {
			log.Fatal(err)
		}

		meta, md, err := extractMetadata(md)
		if err != nil {
			log.Fatal(err)
		}

		posts = append(posts, post{
			Title: meta.Title,
			Path:  strings.Replace(postFile.Name(), ".md", "", 1),
			Date:  meta.Date,
		})

		p := parser.NewWithExtensions(
			parser.CommonExtensions |
				parser.AutoHeadingIDs |
				parser.NoEmptyLineBeforeBlock |
				parser.Footnotes)
		doc := p.Parse(md)

		html := markdown.Render(doc, mdhtml.NewRenderer(mdhtml.RendererOptions{
			Flags:          mdhtml.CommonFlags | mdhtml.HrefTargetBlank,
			RenderNodeHook: mdToHtmlRenderHook,
		}))

		out := filepath.Join(genDir, strings.Replace(postFile.Name(), "md", "html", 1))
		templ := makePostTemplate(template.Must(template.ParseFiles(
			"templates/post.html",
			"templates/layout.html",
			"templates/layout.css")))
		args := struct {
			Content     string
			CurrentYear int
			Date        string
			Title       string
		}{
			Content:     string(html),
			CurrentYear: time.Now().Year(),
			Date:        meta.Date.Format("Jan 02, 2006"),
			Title:       meta.Title,
		}
		if err := r.renderHtml(out, templ, args); err != nil {
			log.Fatal(err)
		}
	}
	sort.Slice(posts, func(i, j int) bool { return posts[i].Date.After(posts[j].Date) })

	// index.html
	{
		out := filepath.Join(genDir, "index.html")
		templ := template.Must(template.ParseFiles(
			"templates/layout.html", "templates/index.html"))
		args := struct {
			CurrentYear int
			Posts       []post
		}{
			CurrentYear: time.Now().Year(),
			Posts:       posts,
		}
		if err := r.renderHtml(out, templ, args); err != nil {
			log.Fatal(err)
		}
	}

	// layout.css
	{
		var codeHighlightStyle strings.Builder
		htmlCodeFormatter.WriteCSS(&codeHighlightStyle, highlightStyle)

		out := filepath.Join(genDir, "layout.css")
		templ := template.Must(template.ParseFiles(
			"templates/layout.css"))
		args := struct {
			CodeHighlightStyle string
		}{
			CodeHighlightStyle: codeHighlightStyle.String(),
		}
		if err := r.renderCss(out, templ, args); err != nil {
			log.Fatal(err)
		}
	}

	// dark-mode.js
	{
		out := filepath.Join(genDir, "dark-mode.js")
		templ := template.Must(template.ParseFiles(
			"templates/dark-mode.js"))
		if err := r.renderJs(out, templ, struct{}{}); err != nil {
			log.Fatal(err)
		}
	}
}
