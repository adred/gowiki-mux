package main

import (
    "fmt"
    "html/template"
    "io/ioutil"
    "net/http"
    "github.com/gorilla/mux"
    "regexp"
)

type Page struct {
    Title string
    Body  []byte
}

func (p *Page) save() error {
    filename := p.Title + ".txt"
    return ioutil.WriteFile(dataDir + filename, p.Body, 0600)
}

func LoadPage(title string) (*Page, error) {
    filename := title + ".txt"
    body, err := ioutil.ReadFile(dataDir + filename)
    if err != nil {
        return nil, err
    }
    return &Page{Title: title, Body: body}, nil
}

func SaveHandler(w http.ResponseWriter, r *http.Request, title string) {
    body := r.FormValue("body")
    p := &Page{Title: title, Body: []byte(body)}
    err := p.save()
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    http.Redirect(w, r, "/view/"+title, http.StatusFound)
}

func ViewHandler(w http.ResponseWriter, r *http.Request, title string) {
    p, err := LoadPage(title)
    fmt.Println(err)
    if err != nil {
        http.Redirect(w, r, "/edit/"+title, http.StatusFound)
        return
    }
    RenderTemplate(w, "view", p)
}

func EditHandler(w http.ResponseWriter, r *http.Request, title string) {
    p, err := LoadPage(title)
    if err != nil {
        p = &Page{Title: title}
    }
    RenderTemplate(w, "edit", p)
}

func makeHandler(fn func(http.ResponseWriter, *http.Request, string)) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        title := mux.Vars(r)["title"]
        if !titleValidator.MatchString(title) {
            http.NotFound(w, r)
            return
        }
        fn(w, r, title)
    }
}

var titleValidator = regexp.MustCompile("^[a-zA-Z0-9]+$")
const templateDir, dataDir = "templates/", "data/"
var templates = template.Must(template.ParseFiles(templateDir + "edit.html", templateDir + "view.html"))

func RenderTemplate(w http.ResponseWriter, tmpl string, p *Page) {
    err := templates.ExecuteTemplate(w, tmpl + ".html", p)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
}

func main() {
    r := mux.NewRouter()
    r.HandleFunc("/view/{title}", makeHandler(ViewHandler))
    r.HandleFunc("/edit/{title}", makeHandler(EditHandler))
    r.HandleFunc("/save/{title}", makeHandler(SaveHandler))
    http.Handle("/", r)
    http.ListenAndServe(":8080", nil)
}