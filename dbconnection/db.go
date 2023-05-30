package dbconnection

import (
	"database/sql"
	"fmt"
)

func SetupDB() (*sql.DB, error) {
	db, err := sql.Open("mysql", "root:Password@tcp(localhost:3306)/crud")

	err = db.Ping()
	if err != nil {
		// os.Exit(1)
		fmt.Print(err)
		return db, err
	}

	return db, err
}
