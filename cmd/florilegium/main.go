package main

import (
	"blekksprut.net/florilegium"
	"embed"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"net/http/cgi"
	"net/url"
	"os"
)

//go:embed *.html
var html embed.FS

//go:embed static/*
var static embed.FS

var templates *template.Template

var (
	Source = "wiki"
	Static = "/static"
	Home   = "florilegium"
)

func redirect(w http.ResponseWriter, r *http.Request) {
	home := os.Getenv("FLORILEGIUM_HOME")
	http.Redirect(w, r, "/"+home, 302)
}

type Server struct {
	Garden *florilegium.Garden
}

type Page struct {
	Name string
	Type string
}

func (p *Page) Static() string {
	return os.Getenv("FLORILEGIUM_STATIC")
}

func (s *Server) List(w http.ResponseWriter, r *http.Request) {
	page := Page{"検索", "list"}
	templates.ExecuteTemplate(w, "header", &page)
	fmt.Fprintln(w, "<nav>")
	fmt.Fprintln(w, "<ul>")
	s.Garden.Stroll(func(name string) {
		fmt.Fprintf(w, "<li><a href='%s'>%s</a>\n", name, name)
	})
	fmt.Fprintln(w, "</ul>")
	fmt.Fprintln(w, "</nav>")
	templates.ExecuteTemplate(w, "footer", &page)
}

func (s *Server) Get(w http.ResponseWriter, r *http.Request) {
	home := os.Getenv("FLORILEGIUM_HOME")
	name := r.PathValue("name")
	if name == "" {
		http.Redirect(w, r, "/"+home, 302)
		return
	}

	if name[len(name)-1] == '/' {
		http.Redirect(w, r, home, 302)
	}

	// if !s.Garden.Exists(name) { redirect }
	page := Page{name, "show"}
	templates.ExecuteTemplate(w, "header", &page)
	fmt.Fprintln(w, "<article>")
	s.Garden.Read(name, w)
	fmt.Fprintln(w, "</article>")
	templates.ExecuteTemplate(w, "footer", &page)
}

func (s *Server) Art(w http.ResponseWriter, r *http.Request) {
	page := Page{"映像", "art"}
	templates.ExecuteTemplate(w, "header", &page)
	templates.ExecuteTemplate(w, "art", &page)

	fmt.Fprintln(w, "<ul class=art>")
	s.Garden.ArtStroll(func(name string) {
		fmt.Fprintln(w, "<li>")
		fmt.Fprintf(w, "<a href='/src/%s'><img src='/t/%s' alt></a>\n", name, name)
	})
	fmt.Fprintln(w, "</ul>")

	templates.ExecuteTemplate(w, "footer", &page)
}

func (s *Server) Upload(w http.ResponseWriter, r *http.Request) {
	f, _, err := r.FormFile("upload")
	if err == nil {
		err = s.Garden.Store(f)
		if err != nil {
			http.Error(w, "file trouble "+err.Error(), 500)
			return
		}
	}
	http.Redirect(w, r, "/a", http.StatusFound)
}

func (s *Server) Edit(w http.ResponseWriter, r *http.Request) {
	name := r.PathValue("name")
	page := Page{name, "edit"}
	templates.ExecuteTemplate(w, "header", &page)

	action := url.QueryEscape(name)
	fmt.Fprintf(w, "<form method=post action='/%s'>", action)
	fmt.Fprintf(w, "<p><textarea name=text cols=72 rows=16 placeholder='…'>")
	s.Garden.Raw(name, w)
	fmt.Fprintln(w, "</textarea>")
	fmt.Fprintln(w, "<input type=submit value=保存>")
	fmt.Fprintln(w, "<input type=reset value=リセット>")
	fmt.Fprintln(w, "</form>")
	templates.ExecuteTemplate(w, "footer", &page)
}

func (s *Server) Post(w http.ResponseWriter, r *http.Request) {
	name := r.PathValue("name")
	text := r.PostFormValue("text")
	err := s.Garden.Plant(name, text)
	if err != nil {
		panic(err)
	}
	http.Redirect(w, r, "/"+name, 302)
}

func main() {
	florilegium.SetEnvIfMissing("FLORILEGIUM_SOURCE", Source)
	florilegium.SetEnvIfMissing("FLORILEGIUM_STATIC", Static)
	florilegium.SetEnvIfMissing("FLORILEGIUM_HOME", Home)

	source := os.Getenv("FLORILEGIUM_SOURCE")
	Lockdown(source)

	root, err := os.OpenRoot(source)
	if err != nil {
		panic(err)
	}

	templates = template.Must(template.ParseFS(html, "*.html"))

	server := Server{
		Garden: &florilegium.Garden{Root: root},
	}

	prefix := ""

	mux := http.NewServeMux()
	mux.HandleFunc("GET "+prefix+"/*", server.List)
	mux.HandleFunc("GET "+prefix+"/a", server.Art)
	mux.HandleFunc("GET "+prefix+"/{name...}", server.Get)
	mux.HandleFunc("GET "+prefix+"/e/{name...}", server.Edit)

	mux.HandleFunc("POST "+prefix+"/a", server.Upload)
	mux.HandleFunc("POST "+prefix+"/{name...}", server.Post)

	if os.Getenv("GATEWAY_INTERFACE") != "" {
		err := cgi.Serve(mux)
		if err != nil {
			panic(err)
		}
		return
	}

	mux.Handle("GET /static/", http.FileServer(http.FS(static)))
	mux.Handle("GET /src/", http.FileServer(http.Dir("")))
	mux.Handle("GET /t/", http.FileServer(http.Dir("")))

	fmt.Fprintln(os.Stderr, "let's listen to :8080")
	log.Fatal(http.ListenAndServe(":8080", mux))
}
