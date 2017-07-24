package main

import (
	"html/template"
	"io/ioutil"
	"net/http"
	"regexp"
	"path/filepath"
	"strings"
	"fmt"
)

type Page struct {
	Title string
	Body  []byte
}

type rootPage struct {
	Title string
	Body  []byte
	Files template.HTML
}


var (
	templatesPath = "template/"
	dataPath = "data/"
)

func (p *Page) save() error {
	filename := dataPath + p.Title + ".txt"
	return ioutil.WriteFile(filename, p.Body, 0600)
}

func loadPage(title string) (*Page, error) {
		filename := dataPath + title + ".txt"
		body, err := ioutil.ReadFile(filename)
		if err != nil {
			return nil, err
		}
		return &Page{Title: title, Body: body}, nil
}

func viewHandler(w http.ResponseWriter, r *http.Request, title string) {
	p, err := loadPage(title)
	if err != nil {
		http.Redirect(w, r, "/edit/"+title, http.StatusFound)
		return
	}
	renderTemplate(w, "view", p)
}

func rootHandler(w http.ResponseWriter, r *http.Request) {
	var a string
	var til string
	//var startHref stringa
	startHref := `<a href=/edit/`
	var endHref	= `</a>`
	til = "root"
	files, _ := ioutil.ReadDir("data/")
	for _, f := range files {
		m := strings.Replace( f.Name(), ".txt", "", 1)
		a = a + startHref + m + `>` + m + endHref + "\n"
	}
	filename := templatesPath + "root.txt"
	body, err := ioutil.ReadFile(filename)
	if err != nil {
	}
	b :=template.HTML(a)
	renderRoot(w, "root", &rootPage{Title: til, Body: body, Files: b})

	fmt.Println(string(a))
}

func editHandler(w http.ResponseWriter, r *http.Request, title string) {
	p, err := loadPage(title)
	if err != nil {
		p = &Page{Title: title}
	}
	renderTemplate(w, "edit", p)
}

func saveHandler(w http.ResponseWriter, r *http.Request, title string) {
	body := r.FormValue("body")
	p := &Page{Title: title, Body: []byte(body)}
	err := p.save()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/view/"+title, http.StatusFound)
}

var templates = template.Must(template.ParseGlob(filepath.Join(templatesPath, "*.html")))

func renderTemplate(w http.ResponseWriter, tmpl string, p *Page) {
	err := templates.ExecuteTemplate(w, tmpl+".html", p)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
func renderRoot(w http.ResponseWriter, tmpl string, p *rootPage) {
	err := templates.ExecuteTemplate(w, "root.html", p)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

var validPath = regexp.MustCompile("^/(edit|save|view)/([a-zA-Z0-9]+)$")

func makeHandler(fn func(http.ResponseWriter, *http.Request, string)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		m := validPath.FindStringSubmatch(r.URL.Path)
		if m == nil {
			http.NotFound(w, r)
			return
		}
		fn(w, r, m[2])
	}
}

func main() {
	http.HandleFunc("/", rootHandler)
	http.HandleFunc("/view/", makeHandler(viewHandler))
	http.HandleFunc("/edit/", makeHandler(editHandler))
	http.HandleFunc("/save/", makeHandler(saveHandler))

	http.ListenAndServe(":8080", nil)
}