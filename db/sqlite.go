package db

import (
	"database/sql"

	_ "github.com/mattn/go-sqlite3"

	"gamestreambot/utils"
)

// TODO: make some functions methods
// TODO: add a way of deleting streams by specifying a list of IDs

type Stream struct {
	ID          int
	Name        string
	Platform    string
	Date        string
	Time        string
	Description string
	URL         string
}

type Streams struct {
	Streams []Stream
}

// create the db with a streams table containing stream information and an options table
// containing server specific options
func CreateDB() error {
	db, openErr := sql.Open("sqlite3", utils.DBFile)
	if openErr != nil {
		return openErr
	}
	defer db.Close()

	sqlStmt := `
	create table if not exists streams (id integer not null primary key, name text, platform text, date text, time text, description text, url text);
	`
	_, tableErr := db.Exec(sqlStmt)
	if tableErr != nil {
		return tableErr
	}
	sqlStmt = `
	create table if not exists settings (server_id integer not null primary key, announce_channel text, announce_role text, playstation boolean, xbox boolean, nintendo boolean, pc boolean, awards boolean);
	`
	_, tableErr = db.Exec(sqlStmt)
	if tableErr != nil {
		return tableErr
	}
	return nil
}
