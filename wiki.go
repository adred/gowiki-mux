package main

import (
    "fmt"
    "strconv"
    "html/template"
    "net/http"
    "github.com/gorilla/mux"
    "regexp"
    "database/sql"
    _ "github.com/go-sql-driver/mysql"
    "os"
)

type Page struct {
    ID    int
    Title string
    Body  string
}

func (p *Page) Add() error {
    st, err := db.Prepare("INSERT INTO page VALUES (?, ?, ?)")
    defer st.Close()
    if err != nil{
        fmt.Print( err )
        os.Exit(1)
    }
    st.Exec(p.ID, p.Title, p.Body)
    return nil
}

func (p *Page) Update() error {
    st, err := db.Prepare("UPDATE page SET title = ?, body = ? WHERE id = ?")
    defer st.Close()
    if err != nil{
        fmt.Print( err )
        os.Exit(1)
    }
    st.Exec(p.Title, p.Body, p.ID)
    return nil
}

func LoadPage(id int) (*Page, error) {
    st, err := db.Prepare("SELECT title, body FROM page WHERE id = ?")
        if err != nil {
        fmt.Print( err )
        os.Exit(1)
    }
    defer st.Close()
    rows, err := st.Query(id)
    if err != nil {
        fmt.Print( err )
        os.Exit(1)
    }
    var title, body string
    for rows.Next() {
        err = rows.Scan( &title, &body )
    }
    return &Page{ID: id, Title: title, Body: body}, nil
}

func AddHandler(w http.ResponseWriter, r *http.Request) {
    st, err := db.Prepare("SELECT id FROM page ORDER BY id DESC LIMIT 1")
        if err != nil {
        fmt.Print( err )
        os.Exit(1)
    }
    defer st.Close()
    var id int
    err = st.QueryRow().Scan(&id)
    id += 1

    p := &Page{ID: id}
    RenderTemplate(w, "add", p)
}

func SaveHandler(w http.ResponseWriter, r *http.Request, id int) {
    title := r.FormValue("title")
    body := r.FormValue("body")
    action := r.FormValue("action")
    p := &Page{ID: id, Title: title, Body: body}
    if action == "add" {
        err := p.Add()
        if err != nil {
            http.Error(w, err.Error(), http.StatusInternalServerError)
            return
        }
    } else if action == "update" {
        err := p.Update()
        if err != nil {
            http.Error(w, err.Error(), http.StatusInternalServerError)
            return
        }
    }
    http.Redirect(w, r, "/view/"+strconv.Itoa(id), http.StatusFound)
}

func ViewHandler(w http.ResponseWriter, r *http.Request, id int) {
    p, err := LoadPage(id)
    if err != nil {
        http.Redirect(w, r, "/edit/"+strconv.Itoa(id), http.StatusFound)
        return
    }
    RenderTemplate(w, "view", p)
}

func EditHandler(w http.ResponseWriter, r *http.Request, id int) {
    p, err := LoadPage(id)
    if err != nil {
        p = &Page{}
    }
    RenderTemplate(w, "edit", p)
}

func MakeHandler(fn func(http.ResponseWriter, *http.Request, int)) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        if !idValidator.MatchString(mux.Vars(r)["id"]) {
            http.NotFound(w, r)
            return
        }
        id, _ := strconv.Atoi(mux.Vars(r)["id"])
        fn(w, r, id)
    }
}

func RenderTemplate(w http.ResponseWriter, tmpl string, p *Page) {
    err := templates.ExecuteTemplate(w, tmpl + ".html", p)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
}

func OpenDB(driver string, username string, password string, dbname string) {
    var err error
    db, err = sql.Open(driver, username + ":" + password + "@/" + dbname)
    if err != nil {
        fmt.Print( err )
        os.Exit(1)
    }
    err = db.Ping()
    if err != nil {
        fmt.Print( err )
        os.Exit(1)
    }
}

var idValidator = regexp.MustCompile("^[0-9]+$")
const templateDir, dataDir = "templates/", "data/"
var templates = template.Must(template.ParseFiles(templateDir + "add.html", templateDir + "edit.html", templateDir + "view.html"))
var db *sql.DB

func main() {
    OpenDB("mysql", "root", "redeye", "gowiki")

    r := mux.NewRouter()
    r.HandleFunc("/add", AddHandler)
    r.HandleFunc("/view/{id}", MakeHandler(ViewHandler))
    r.HandleFunc("/edit/{id}", MakeHandler(EditHandler))
    r.HandleFunc("/save/{id}", MakeHandler(SaveHandler))
    http.Handle("/", r)
    http.ListenAndServe(":8080", nil)
}