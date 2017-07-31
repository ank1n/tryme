package main

import (
	"database/sql"
	"fmt"
	"os"
	"os/user"
	"strconv"
	"unicode/utf8"

	"github.com/mini"
	_ "github.com/lib/pq"
)

const help = `Usage: phonebook COMMAND [ARG]...
Commands:
	add NAME PHONE - create new record;
	del ID1 ID2... - delete record;
	edit ID        - edit record;
	show           - display all records;
	show STRING    - display records which contain a given substring in the name;
	help           - display this help.`

type record struct {
	id          int
	name, phone string
}

func insert(db *sql.DB, name, phone string) (int64, error) {
	res, err := db.Exec("INSERT INTO phonebook VALUES (default, $1, $2)", name, phone)
	if err != nil {
		return 0, err
	}
	return res.RowsAffected()
}

func remove(db *sql.DB, ids []string) error {
	stmt, err := db.Prepare("DELETE FROM phonebook WHERE id=$1")
	if err != nil {
		return err
	}
	defer stmt.Close()
	for _, v := range ids {
		_, err := stmt.Exec(v)
		if err != nil {
			return err
		}
	}
	return nil
}

func update(db *sql.DB, id, name, phone string) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	_, err = tx.Exec("UPDATE phonebook SET name = $1, phone = $2 WHERE id=$3",
		name, phone, id)
	if err != nil {
		return err
	}
	return tx.Commit()
}

func show(db *sql.DB, arg string) ([]record, error) {
	var s string
	if len(arg) != 0 {
		s = "WHERE name LIKE '%" + arg + "%'"
	}
	rows, err := db.Query("SELECT * FROM phonebook " + s + " ORDER BY id")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var rs = make([]record, 0)
	var rec record
	for rows.Next() {
		err = rows.Scan(&rec.id, &rec.name, &rec.phone)
		if err != nil {
			return nil, err
		}
		rs = append(rs, rec)
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

func format(rs []record) {
	var max, tmp int
	for _, v := range rs {
		tmp = utf8.RuneCountInString(v.name)
		if max < tmp {
			max = tmp
		}
	}
	s := "%-" + strconv.Itoa(max) + "s"
	for _, v := range rs {
		fmt.Printf("%3d   "+s+"   %s\n", v.id, v.name, v.phone)
	}
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

func main() {
	if len(os.Args) < 2 {
		fatal("Usage: phonebook COMMAND [ARG]...")
	}
	db, err := sql.Open("postgres", params())
	chk(err)
	defer db.Close()

	_, err = db.Exec("CREATE TABLE IF NOT EXISTS " +
		`phonebook("id" SERIAL PRIMARY KEY,` +
		`"name" varchar(50), "phone" varchar(100))`)
	chk(err)

	switch os.Args[1] {
	case "add":
		if len(os.Args) != 4 {
			fatal("Usage: phonebook add NAME PHONE")
		}
		num, err := insert(db, os.Args[2], os.Args[3])
		chk(err)
		fmt.Println(num, "rows affected")
	case "del":
		if len(os.Args) < 3 {
			fatal("Usage: phonebook del ID...")
		}
		}
		err = remove(db, os.Args[2:])
		chk(err)
	case "edit":
		if len(os.Args) != 5 {
			fatal("Usage: phonebook edit ID NAME PHONE")
		}
		err = update(db, os.Args[2], os.Args[3], os.Args[4])
		chk(err)
	case "show":
		if len(os.Args) > 3 {
			fatal("Usage: phonebook show [SUBSTRING]")
		}
		var s string
		if len(os.Args) == 3 {
			s = os.Args[2]
		}
		res, err := show(db, s)
		chk(err)
		format(res)
	case "help":
		fmt.Println(help)
	default:
		fatal("Invalid command: " + os.Args[1])
	}
}
