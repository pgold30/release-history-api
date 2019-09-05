package db

import (
	"database/sql"

	_ "github.com/lib/pq" // importing postgres driver only for its initialization
)

func GetDB(connString string) *sql.DB {
	db, err := sql.Open("postgres", connString)
	if err != nil {
		panic(err)
	}
	// test connection
	err = db.Ping()
	if err != nil {
		panic(err)
	}
	runQuery(db, `
		CREATE TABLE IF NOT EXISTS deployment
		(
			d_id          SERIAL,
			d_project     TEXT                     NOT NULL,
			d_service     TEXT                     NOT NULL,
			d_environment TEXT                     NOT NULL,
			d_tag         TEXT                     NOT NULL,
			d_date        TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT clock_timestamp(),
			CONSTRAINT pk_deployment PRIMARY KEY (d_id)
		)
	`)
	runQuery(db, `
		CREATE TABLE IF NOT EXISTS release
		(
			r_id          	SERIAL,
			r_project     	TEXT                     NOT NULL,
			r_number     	TEXT                     NOT NULL,
			r_date        	TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT clock_timestamp(),
			CONSTRAINT pk_release PRIMARY KEY (r_id)
		)
	`)
	return db
}

func runQuery(db *sql.DB, query string) {
	stmt, err := db.Prepare(query)
	if err != nil {
		panic(err)
	}
	_, err = stmt.Exec()
	if err != nil {
		panic(err)
	}
}
