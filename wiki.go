package main

import (
    "fmt"
    "html/template"
    "net/http"
    "github.com/gorilla/mux"
    "regexp"
    "database/sql"
    _ "github.com/go-sql-driver/mysql"
    "os"
)

type Page struct {
    ID    string
    Title string
    Body  string
}

func (p *Page) Save() error {
    st, err := db.Prepare("INSERT INTO page VALUES (?, ?, ?)")
    defer st.Close()
    if err != nil{
        fmt.Print( err )
        os.Exit(1)
    }
    st.Exec(p.ID, p.Title, p.Body)
    return nil
}

func LoadPage(id string) (*Page, error) {
    st, err := db.Prepare("SELECT title, body FROM page WHERE id = ?")
        if err != nil{
        fmt.Print( err );
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

func SaveHandler(w http.ResponseWriter, r *http.Request, id string) {
    title := r.FormValue("title")
    body := r.FormValue("body")
    p := &Page{ID: id, Title: title, Body: body}
    err := p.Save()
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    http.Redirect(w, r, "/view/"+id, http.StatusFound)
}

func ViewHandler(w http.ResponseWriter, r *http.Request, id string) {
    p, err := LoadPage(id)
    fmt.Println(err)
    if err != nil {
        http.Redirect(w, r, "/edit/"+id, http.StatusFound)
        return
    }
    RenderTemplate(w, "view", p)
}

func EditHandler(w http.ResponseWriter, r *http.Request, id string) {
    p, err := LoadPage(id)
    if err != nil {
        p = &Page{}
    }
    RenderTemplate(w, "edit", p)
}

func makeHandler(fn func(http.ResponseWriter, *http.Request, string)) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        id := mux.Vars(r)["id"]
        if !idValidator.MatchString(id) {
            http.NotFound(w, r)
            return
        }
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
var templates = template.Must(template.ParseFiles(templateDir + "edit.html", templateDir + "view.html"))
var db *sql.DB

func main() {
    OpenDB("mysql", "root", "redeye", "gowiki")

    r := mux.NewRouter()
    r.HandleFunc("/view/{id}", makeHandler(ViewHandler))
    r.HandleFunc("/edit/{id}", makeHandler(EditHandler))
    r.HandleFunc("/save/{id}", makeHandler(SaveHandler))
    http.Handle("/", r)
    http.ListenAndServe(":8080", nil)
}