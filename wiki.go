package main

import (
    "fmt"
    "os"
    "strconv"
    "html/template"
    "net/http"
    "github.com/gorilla/mux"
    "database/sql"
    _ "github.com/go-sql-driver/mysql"
)

type Page struct {
    ID    int
    Title string
    Body  string
}

func (p *Page) Add() int {
    st, err := db.Prepare("INSERT INTO page VALUES ('', ?, ?)")
    defer st.Close()
    if err != nil{
        fmt.Print( err )
        os.Exit(1)
    }
    res, err := st.Exec(p.Title, p.Body)
    if err != nil{
        fmt.Print( err )
        os.Exit(1)
    }
    id, err := res.LastInsertId()
    if err != nil{
        fmt.Print( err )
        os.Exit(1)
    }
    idFinal, _ := strconv.Atoi(strconv.FormatInt(id, 10))

    return idFinal
}

func (p *Page) Update() error {
    st, err := db.Prepare("UPDATE page SET title = ?, body = ? WHERE id = ?")
    defer st.Close()
    if err != nil{
        fmt.Print( err )
        os.Exit(1)
    }
    res, err := st.Exec(p.Title, p.Body, p.ID)
    if err != nil{
        fmt.Print( err )
        os.Exit(1)
    }
    _, err = res.RowsAffected()
    if err != nil{
        fmt.Print( err )
        os.Exit(1)
    }

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
    p := &Page{}
    RenderTemplate(w, "add", p)
}

func SaveHandler(w http.ResponseWriter, r *http.Request, id int) {
    title := r.FormValue("title")
    body := r.FormValue("body")
    action := r.FormValue("action")
    if action == "add" {
        p := &Page{Title: title, Body: body}
        id = p.Add()
    } else if action == "update" {
        p := &Page{ID: id, Title: title, Body: body}
        err := p.Update()
        if err != nil {
            http.Redirect(w, r, "/edit/"+strconv.Itoa(id), http.StatusFound)
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
        p = &Page{ID: id}
    }
    RenderTemplate(w, "edit", p)
}

func MakeHandler(fn func(http.ResponseWriter, *http.Request, int)) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
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

const templateDir, dataDir = "templates/", "data/"
var templates = template.Must(template.ParseFiles(templateDir + "add.html", templateDir + "edit.html", templateDir + "view.html"))
var db *sql.DB

func main() {
    OpenDB("mysql", "root", "redeye", "gowiki")

    r := mux.NewRouter()
    r.HandleFunc("/add", AddHandler)
    r.HandleFunc("/view/{id:[0-9]+}", MakeHandler(ViewHandler))
    r.HandleFunc("/edit/{id:[0-9]+}", MakeHandler(EditHandler))
    r.HandleFunc("/save/{id:[0-9]+}", MakeHandler(SaveHandler))
    r.HandleFunc("/save", MakeHandler(SaveHandler))
    http.Handle("/", r)
    http.ListenAndServe(":8080", nil)
}