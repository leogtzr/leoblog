package main

import (
	"bytes"
	"github.com/yuin/goldmark"
	highlighting "github.com/yuin/goldmark-highlighting/v2"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
)

type SlugReader interface {
	Read(slug string) (string, error)
}

type FileReader struct {
}

type PostData struct {
	Title   string
	Content template.HTML
	Author  string
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
		slug := r.PathValue("slug")
		postMarkdown, err := sl.Read(slug)
		if err != nil {
			// TODO: handle error
			log.Printf("error: %v", err)
			http.Error(w, "Post not found", http.StatusNotFound)
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
		err = mdRenderer.Convert([]byte(postMarkdown), &buf)
		if err != nil {
			log.Printf("error: %v", err)
			http.Error(w, "Error converting markdown", http.StatusInternalServerError)
			return
		}

		tpl, err := template.ParseFiles("post.gohtml")
		if err != nil {
			http.Error(w, "Error parsing template", http.StatusInternalServerError)
			return
		}

		err = tpl.Execute(w, PostData{
			Title:   "My First Post",
			Content: template.HTML(buf.String()),
			Author:  "Leo Gutiérrez",
		})

		if err != nil {
			log.Printf("error: %v", err)
		}
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
