package db

import (
	"database/sql"

	_ "github.com/mattn/go-sqlite3"

	"gamestreambot/utils"
)

// create the db with a streams table containing stream information and an options table
// containing server specific options
func CreateDB() error {
	utils.Log.Info.WithPrefix(" MAIN").Info("loading/creating database")
	db, openErr := sql.Open("sqlite3", utils.Files.DB)
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
	create table if not exists config (id integer not null primary key, stream_url text, api_url text, last_updated text);
	`
	_, tableErr = db.Exec(sqlStmt)
	if tableErr != nil {
		return tableErr
	}
	sqlStmt = `
	create table if not exists servers (server_id integer not null primary key, announce_channel text, announce_role text, playstation boolean, xbox boolean, nintendo boolean, pc boolean);
	`
	_, tableErr = db.Exec(sqlStmt)
	if tableErr != nil {
		return tableErr
	}
	return nil
}
