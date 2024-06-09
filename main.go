package main

import (
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
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"regexp"
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

func getTitle(doc ast.Node) string {
	var title string
	ast.WalkFunc(doc, func(node ast.Node, entering bool) ast.WalkStatus {
		h, ok := node.(*ast.Heading)
		if !ok {
			return ast.GoToNext
		}
		if h.Level != 1 {
			return ast.GoToNext
		}
		if len(h.Children) != 1 {
			return ast.GoToNext
		}

		t, ok := h.Children[0].(*ast.Text)
		if !ok {
			return ast.GoToNext
		}

		title = string(t.Literal)
		return ast.Terminate
	})
	return title
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

func makeHtmlRenderer() *htmlRenderer {
	p := &htmlRenderer{}

	p.m = minify.New()
	p.m.AddFunc("text/css", css.Minify)
	p.m.Add("text/html", &minifyhtml.Minifier{
		KeepEndTags: true,
	})
	p.m.AddFuncRegexp(
		regexp.MustCompile("^(application|text)/(x-)?(java|ecma)script$"), js.Minify)

	return p
}

func makeLayoutTemplate() *template.Template {
	return template.Must(template.ParseFiles(
		"templates/layout.html",
		"templates/structure.css"))
}

func makePostTemplate(
	layoutTempl *template.Template, content string) *template.Template {

	return template.Must(layoutTempl.New("article").Parse(content)).Lookup("layout.html")
}

func main() {
	layoutTemplate := makeLayoutTemplate()

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

	postFiles := map[string]string{}
	for _, post := range posts {
		in := filepath.Join(postsDir, post.Name())
		md, err := ioutil.ReadFile(in)
		if err != nil {
			log.Fatal(err)
		}

		p := parser.NewWithExtensions(
			parser.CommonExtensions |
				parser.AutoHeadingIDs |
				parser.NoEmptyLineBeforeBlock)
		doc := p.Parse(md)

		postFiles[getTitle(doc)] = strings.Replace(post.Name(), ".md", "", 1)

		mdr := mdhtml.NewRenderer(mdhtml.RendererOptions{
			Flags:          mdhtml.CommonFlags | mdhtml.HrefTargetBlank,
			RenderNodeHook: mdToHtmlRenderHook,
		})
		html := markdown.Render(doc, mdr)

		out := filepath.Join(genDir, strings.Replace(post.Name(), "md", "html", 1))
		args := map[string]string{
			"CodeHighlightStyle": codeHighlightStyle.String(),
			"Content":            string(html),
			"CurrentYear":        strconv.Itoa(time.Now().Year()),
		}

		postTempl := makePostTemplate(
			template.Must(layoutTemplate.Clone()), string(html))

		if err := r.renderHtml(out, postTempl, args); err != nil {
			log.Fatal(err)
		}
	}

	indexOut := filepath.Join(genDir, "index.html")
	indexTempl := template.Must(template.ParseFiles(
		"templates/layout.html",
		"templates/index.html",
		"templates/structure.css"))

	indexArgs := struct {
		CurrentYear        string
		Posts              map[string]string
		CodeHighlightStyle string
	}{
		CurrentYear: strconv.Itoa(time.Now().Year()),
		Posts:       postFiles,
	}
	if err := r.renderHtml(indexOut, indexTempl, indexArgs); err != nil {
		log.Fatal(err)
	}
}
