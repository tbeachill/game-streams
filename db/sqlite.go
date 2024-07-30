package db

import (
	"database/sql"

	_ "github.com/mattn/go-sqlite3"

	"gamestreams/utils"
)

// CreateDB creates the database if it does not exist. It creates the streams, config,
// and servers tables.
// streams contains information about the streams.
// config contains configuration information for the bot.
// servers contains information about the servers that the bot is in and their settings.
// blacklist contains information about users and servers that are blacklisted from
// using the bot.
func CreateDB() error {
	utils.LogInfo(" MAIN", "loading/creating database", false)
	db, openErr := sql.Open("sqlite3", utils.Files.DB)
	if openErr != nil {
		return openErr
	}
	defer db.Close()

	_, tableErr := db.Exec(`CREATE TABLE IF NOT EXISTS streams
								(id INTEGER NOT NULL PRIMARY KEY,
								stream_name TEXT,
								platform TEXT,
								stream_date TEXT,
								start_time TEXT,
								stream_desc TEXT,
								stream_url TEXT)`)

	if tableErr != nil {
		return tableErr
	}

	_, tableErr = db.Exec(`CREATE TABLE IF NOT EXISTS config
								(id INTEGER NOT NULL PRIMARY KEY,
								toml_url TEXT,
								api_url TEXT,
								last_updated TEXT)`)

	if tableErr != nil {
		return tableErr
	}

	_, tableErr = db.Exec(`CREATE TABLE IF NOT EXISTS servers
								(server_id INTEGER NOT NULL PRIMARY KEY,
								server_name TEXT,
								owner_id INTEGER,
								date_joined TEXT,
								member_count INTEGER,
								locale TEXT)`)

	if tableErr != nil {
		return tableErr
	}

	_, tableErr = db.Exec(`CREATE TABLE IF NOT EXISTS server_settings
								(server_id INTEGER NOT NULL PRIMARY KEY,
								announce_channel INTEGER,
								announce_role INTEGER,
								playstation BOOLEAN,
								xbox BOOLEAN,
								nintendo BOOLEAN,
								pc BOOLEAN,
								vr BOOLEAN)`)

	if tableErr != nil {
		return tableErr
	}

	_, tableErr = db.Exec(`CREATE TABLE IF NOT EXISTS blacklist
								(discord_id INTEGER NOT NULL PRIMARY KEY,
								id_type TEXT,
								date_added TEXT,
								date_expires TEXT,
								reason TEXT)`)

	if tableErr != nil {
		return tableErr
	}

	_, tableErr = db.Exec(`CREATE TABLE IF NOT EXISTS commands
								(id INTEGER NOT NULL PRIMARY KEY,
								server_id INTEGER,
								date_time TEXT,
								command TEXT,
								options TEXT,
								response_time_ms INTEGER)`)

	if tableErr != nil {
		return tableErr
	}

	return nil
}
