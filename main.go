package main

import (
	"html/template"
	"net/http"
	"regexp"
	"path/filepath"
	"fmt"
	"os"
	"database/sql"
	"github.com/mini"
	_ "github.com/lib/pq"
	"unicode/utf8"
	"strconv"
	"os/user"
	"io/ioutil"
	"errors"
)

type Page struct {
	Id    int64
	Title string
	Body  []byte
}

type Pages struct {
	Pages []Page
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

func insert(title string, body []byte) (int64, error) {
	db, err := sql.Open("postgres", params())
	chk(err)
	defer db.Close()

	res, err := db.Exec("INSERT INTO notebook VALUES (default, $1, $2)", title, body)
	if err != nil {
		return 0, err
	}
	return res.RowsAffected()
}

func update(id int64, title string, body []byte) error {
	db, err := sql.Open("postgres", params())
	chk(err)
	defer db.Close()
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	_, err = tx.Exec("UPDATE notebook SET title = $1, body = $2 WHERE id=$3", title, body, id)
	if err != nil {
		return err
	}
	return tx.Commit()
}

func show(db *sql.DB, arg string) ([]Page, error) {
	var s string
	if len(arg) != 0 {
		s = "WHERE name LIKE '%" + arg + "%'"
	}
	rows, err := db.Query("SELECT * FROM notebook " + s + " ORDER BY id")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var rs = make([]Page, 0)
	var pa Page
	for rows.Next() {
		err = rows.Scan(&pa.Id, &pa.Title, &pa.Body)
		if err != nil {
			return nil, err
		}
		rs = append(rs, pa)
	}
	err = rows.Err()
	if err != nil {
		return nil, err
	}
	return rs, nil
}

func fatal(v interface{}) {
	fmt.Println(v)
	os.Exit(1)
}

func chk(err error) {
	if err != nil {
		fatal(err)
	}
}

func format(rs []Page) {
	var max, tmp int
	for _, v := range rs {
		tmp = utf8.RuneCountInString(v.Title)
		if max < tmp {
			max = tmp
		}
	}
	s := "%-" + strconv.Itoa(max) + "s"
	for _, v := range rs {
		fmt.Printf("%3d   "+s+"   %s\n", v.Id, v.Title, v.Body)
	}
}

func loadPage(title string) (*Page, error) {

	bks := make([]*Page, 0)
    println("loadpage")
	bks, err := getDb()
	if err != nil {
	}

	for _, bk := range bks {
		if bk.Title == title {
			return &Page{Title: bk.Title, Body: bk.Body}, nil
			fmt.Println(bk.Title)
		} else {
			return nil, nil
		}
	}
	return nil, nil
}

func params() string {
	u, err := user.Current()
	chk(err)

	cfg, err := mini.LoadConfiguration(u.HomeDir + "/go/src/tryme/default.conf")
	chk(err)
	fmt.Printf("configdir=%s", u.HomeDir )

	info := fmt.Sprintf("host=%s port=%s dbname=%s "+
		"sslmode=%s user=%s password=%s ",
		cfg.String("host", "127.0.0.1"),
		cfg.String("port", "5432"),
		cfg.String("dbname", u.Username),
		cfg.String("sslmode", "disable"),
		cfg.String("user", u.Username),
		cfg.String("pass", ""),
	)
	return info
}
func exist(title string) (int64, error) {
	db, err := sql.Open("postgres", params())
	chk(err)
	defer db.Close()
	var name= title
	rows, err := db.Query("SELECT * FROM notebook where name like '" + name + "' ORDER BY id")
	//defer rows.Close();
	if rows == nil {
		return 0, nil
	}
	for rows.Next() {
		var count int64
		println(rows)
		zoo := new(Page)
		rows.Scan(&zoo.Id, &zoo.Title, &zoo.Body)
		println(rows)
		rows.Scan(&count)
		var idR = zoo.Id
		if count > 1 {
			return count, nil
		} else {
			return idR, errors.New("can't work")
		}
	}
	return 0, errors.New("can't work")
}

func getDb() ([]*Page, error) {
	db, err := sql.Open("postgres", params())
	chk(err)
	defer db.Close()

	rows, err := db.Query("SELECT * FROM notebook ORDER BY id")
	bks := make([]*Page, 0)
	for rows.Next() {
		bk := new(Page)
		err := rows.Scan(&bk.Id, &bk.Title, &bk.Body)
		if err != nil {
			return nil, err
		}
		bks = append(bks, bk)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return bks, nil
	defer rows.Close()
	return bks, nil
}

func viewHandler(w http.ResponseWriter, r *http.Request, title string) {
	p, err := loadPage(title)
	println("call viewHandler")
	if err != nil {
		http.Redirect(w, r, "/edit/"+title, http.StatusFound)
		return
	}
	renderTemplate(w, "view", p)
}

func rootHandler(w http.ResponseWriter, r *http.Request) {
	var a string
	var til string
	startHref := `<a href=/edit/`
	var endHref	= `</a>`
	til = "root"
	bks := make([]*Page, 0)

	bks, err := getDb()
	if err != nil {
	}
	for _, bk := range bks {
		a = a + startHref + bk.Title + `>` + bk.Title + endHref + "\n"
	}
	filename := templatesPath + "root.txt"
	body, err := ioutil.ReadFile(filename)
	if err != nil {
	}
	b :=template.HTML(a)
	renderRoot(w, "root", &rootPage{Title: til, Body: body, Files: b})
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
	//p := &Page{Title: title, Body: []byte(body)}
	u, err := exist(title)
	if err == nil {
		update(u, title, []byte(body))
	} else {
		if u == 0 {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		} else {
			insert(title, []byte(body))
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
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
	db, err := sql.Open("postgres", params())
	chk(err)
	defer db.Close()

	_, err = db.Exec("CREATE TABLE IF NOT EXISTS " +
		`notebook("id" SERIAL PRIMARY KEY,` +
		`"title" varchar(50), "body" varchar(100))`)
	chk(err)

	http.HandleFunc("/", rootHandler)
	http.HandleFunc("/view/", makeHandler(viewHandler))
	http.HandleFunc("/edit/", makeHandler(editHandler))
	http.HandleFunc("/save/", makeHandler(saveHandler))

	http.ListenAndServe(":8080", nil)
}