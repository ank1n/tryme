package lib

import "database/sql"

type record struct {
	id          int
	name, phone string
}

func Insert(db *sql.DB, name, phone string) (int64, error) {
	res, err := db.Exec("INSERT INTO phonebook VALUES (default, $1, $2)",
		name, phone)
	if err != nil {
		return 0, err
	}
	return res.RowsAffected()
}

func Remove(db *sql.DB, ids []string) error {
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

func Update(db *sql.DB, id, name, phone string) error {
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

func Show(db *sql.DB, arg string) ([]record, error) {
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
