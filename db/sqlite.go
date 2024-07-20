package db

import (
	"database/sql"

	_ "github.com/mattn/go-sqlite3"

	"gamestreambot/utils"
)

// CreateDB creates the database if it does not exist. It creates the streams, config, and servers tables.
// streams contains information about the streams.
// config contains configuration information for the bot.
// servers contains information about the servers that the bot is in and their settings.
func CreateDB() error {
	utils.Log.Info.WithPrefix(" MAIN").Info("loading/creating database")
	db, openErr := sql.Open("sqlite3", utils.Files.DB)
	if openErr != nil {
		return openErr
	}
	defer db.Close()

	_, tableErr := db.Exec(`CREATE TABLE IF NOT EXISTS streams
								(id INTEGER NOT NULL PRIMARY KEY,
								name TEXT,
								platform TEXT,
								date TEXT,
								time TEXT,
								description TEXT,
								url TEXT)`)

	if tableErr != nil {
		return tableErr
	}

	_, tableErr = db.Exec(`CREATE TABLE IF NOT EXISTS config
								(id INTEGER NOT NULL PRIMARY KEY,
								stream_url TEXT,
								api_url TEXT,
								last_updated TEXT)`)

	if tableErr != nil {
		return tableErr
	}

	_, tableErr = db.Exec(`CREATE TABLE IF NOT EXISTS servers
								(server_id INTEGER NOT NULL PRIMARY KEY,
								announce_channel TEXT,
								announce_role TEXT,
								playstation BOOLEAN,
								xbox BOOLEAN,
								nintendo BOOLEAN,
								pc BOOLEAN,
								vr BOOLEAN);`)

	if tableErr != nil {
		return tableErr
	}
	return nil
}
