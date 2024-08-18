package db

import (
	"database/sql"

	_ "github.com/mattn/go-sqlite3"

	"gamestreams/config"
	"gamestreams/logs"
)

// CreateDB creates the database if it does not exist. It creates the streams, config,
// and servers tables.
// streams contains information about the streams.
// streams_toml contains information about updating the streams table from a toml file.
// servers contains information about the servers that the bot is in.
// server_settings contains the settings for each server.
// blacklist contains information about users and servers that are blacklisted from
// using the bot.
// commands contains information about commands that are run by users.
// suggestions contains information about stream suggestions that are made by users.
// suggestions_archive contains anonymised suggestions for later use.
func CreateDB() error {
	logs.LogInfo(" MAIN", "loading/creating database", false)
	db, openErr := sql.Open("sqlite3", config.Values.Files.Database+"?_fk=1&_cache_size=10000")
	if openErr != nil {
		return openErr
	}
	defer db.Close()

	_, tableErr := db.Exec(`CREATE TABLE IF NOT EXISTS streams
								(id INTEGER PRIMARY KEY AUTOINCREMENT,
								stream_name TEXT,
								platform TEXT,
								stream_date TEXT,
								start_time TEXT,
								stream_desc TEXT,
								stream_url TEXT)`)

	if tableErr != nil {
		return tableErr
	}

	_, tableErr = db.Exec(`CREATE TABLE IF NOT EXISTS stream_toml
								(id INTEGER PRIMARY KEY AUTOINCREMENT,
								last_updated TEXT)`)

	if tableErr != nil {
		return tableErr
	}

	_, tableErr = db.Exec(`CREATE TABLE IF NOT EXISTS servers
								(server_id TEXT NOT NULL PRIMARY KEY,
								server_name TEXT,
								owner_id TEXT,
								date_joined TEXT,
								member_count INTEGER,
								locale TEXT)`)

	if tableErr != nil {
		return tableErr
	}

	_, tableErr = db.Exec(`CREATE TABLE IF NOT EXISTS server_settings
								(server_id TEXT NOT NULL PRIMARY KEY,
								announce_channel TEXT,
								announce_role TEXT,
								playstation BOOLEAN,
								xbox BOOLEAN,
								nintendo BOOLEAN,
								pc BOOLEAN,
								vr BOOLEAN,
								FOREIGN KEY (server_id) REFERENCES servers (server_id)
									ON DELETE CASCADE)`)

	if tableErr != nil {
		return tableErr
	}

	_, tableErr = db.Exec(`CREATE TABLE IF NOT EXISTS blacklist
								(discord_id TEXT NOT NULL,
								id_type TEXT,
								date_added TEXT,
								date_expires TEXT NOT NULL,
								reason TEXT,
								last_messaged TEXT,
								PRIMARY KEY (discord_id, date_expires))`)

	if tableErr != nil {
		return tableErr
	}

	_, tableErr = db.Exec(`CREATE TABLE IF NOT EXISTS commands
								(id INTEGER PRIMARY KEY AUTOINCREMENT,
								server_id TEXT,
								user_id TEXT,
								used_date TEXT,
								used_time TEXT,
								command TEXT,
								options TEXT,
								response_time_ms INTEGER,
								FOREIGN KEY (server_id) REFERENCES servers (server_id)
									ON DELETE CASCADE)`)

	if tableErr != nil {
		return tableErr
	}

	_, tableErr = db.Exec(`CREATE TABLE IF NOT EXISTS suggestions
								(id INTEGER PRIMARY KEY AUTOINCREMENT,
								command_id INTEGER,
								stream_name TEXT,
								stream_date TEXT,
								stream_url TEXT,
								FOREIGN KEY (command_id) REFERENCES commands (id)
									ON DELETE CASCADE
									ON UPDATE CASCADE)`)

	if tableErr != nil {
		return tableErr
	}

	_, tableErr = db.Exec(`CREATE TABLE IF NOT EXISTS suggestions_archive
								(id INTEGER PRIMARY KEY AUTOINCREMENT,
								stream_name TEXT,
								stream_date TEXT,
								stream_url TEXT,
								spam BOOLEAN)`)

	if tableErr != nil {
		return tableErr
	}

	return nil
}
