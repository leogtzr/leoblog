package main

import (
	"bytes"
	"github.com/adrg/frontmatter"
	"github.com/yuin/goldmark"
	highlighting "github.com/yuin/goldmark-highlighting/v2"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
)

type SlugReader interface {
	Read(slug string) (string, error)
}

type FileReader struct {
}

type Post struct {
	Title   string `toml:"title"`
	Slug    string `toml:"slug"`
	Content template.HTML
	Author  Author `toml:"author"`
}

type Author struct {
	Name  string `toml:"name"`
	Email string `toml:"email"`
}

func (fsr FileReader) Read(slug string) (string, error) {
	f, err := os.Open(slug + ".md")
	if err != nil {
		return "", err
	}

	defer f.Close()
	b, err := io.ReadAll(f)
	if err != nil {
		return "", err
	}

	return string(b), nil
}

func PostHandler(sl SlugReader) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var post Post
		post.Slug = r.PathValue("slug")
		postMarkdown, err := sl.Read(post.Slug)
		if err != nil {
			// TODO: handle error
			log.Printf("error: %v", err)
			http.Error(w, "Post not found", http.StatusNotFound)
			return
		}

		postDataResponse, err := frontmatter.Parse(strings.NewReader(postMarkdown), &post)
		if err != nil {
			log.Printf("error: %v", err)
			http.Error(w, "Error parsing frontmatter information", http.StatusInternalServerError)
			return
		}

		mdRenderer := goldmark.New(
			goldmark.WithExtensions(
				highlighting.NewHighlighting(
					highlighting.WithStyle("dracula"),
				),
			),
		)

		var buf bytes.Buffer
		err = mdRenderer.Convert(postDataResponse, &buf)
		if err != nil {
			log.Printf("error: %v", err)
			http.Error(w, "Error converting markdown", http.StatusInternalServerError)
			return
		}

		tpl, err := template.ParseFiles("post.gohtml")
		if err != nil {
			log.Printf("error: %v", err)
			http.Error(w, "Error parsing template", http.StatusInternalServerError)
			return
		}

		post.Content = template.HTML(buf.String())
		err = tpl.Execute(w, post)
	}
}

func main() {
	mux := http.NewServeMux()

	postFileReader := FileReader{}
	mux.HandleFunc("GET /posts/{slug}", PostHandler(postFileReader))

	err := http.ListenAndServe(":3030", mux)
	if err != nil {
		log.Fatal(err)
	}
}
