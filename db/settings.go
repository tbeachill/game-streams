package db

import (
	"database/sql"

	_ "github.com/mattn/go-sqlite3"

	"gamestreambot/utils"
)

type Options struct {
	ServerID        string
	AnnounceChannel string
	AnnounceRole    string
	Playstation     bool
	Xbox            bool
	Nintendo        bool
	PC              bool
	Awards          bool
}

// add a server to the settings table with default options
func SetDefaultOptions(serverID string) error {
	db, openErr := sql.Open("sqlite3", utils.DBFile)
	if openErr != nil {
		return openErr
	}
	defer db.Close()

	_, execErr := db.Exec("insert into settings (server_id, announce_channel, announce_role, playstation, xbox, nintendo, pc, awards) values (?, ?, ?, ?, ?, ?, ?, ?)", serverID, "", "", 0, 0, 0, 0, 0)
	if execErr != nil {
		return execErr
	}
	return nil
}

// remove a server from the settings table
func RemoveOptions(serverID string) error {
	db, openErr := sql.Open("sqlite3", utils.DBFile)
	if openErr != nil {
		return openErr
	}
	defer db.Close()

	_, execErr := db.Exec("delete from settings where server_id = ?", serverID)
	if execErr != nil {
		return execErr
	}
	return nil
}

// set the options for a server from an options struct
func SetOptions(options *Options) error {
	db, openErr := sql.Open("sqlite3", utils.DBFile)
	if openErr != nil {
		return openErr
	}
	defer db.Close()
	checkOptions(options.ServerID)

	_, execErr := db.Exec("update settings set announce_channel = ?, announce_role = ?, playstation = ?, xbox = ?, nintendo = ?, pc = ?, awards = ? where server_id = ?", options.AnnounceChannel, options.AnnounceRole, options.Playstation, options.Xbox, options.Nintendo, options.PC, options.Awards, options.ServerID)
	if execErr != nil {
		return execErr
	}
	return nil
}

// get the options for a server and return as an options struct
func GetOptions(serverID string) (Options, error) {
	db, openErr := sql.Open("sqlite3", utils.DBFile)
	if openErr != nil {
		return Options{}, openErr
	}
	defer db.Close()
	checkOptions(serverID)

	var options Options
	row := db.QueryRow("select * from settings where server_id = ?", serverID)
	scanErr := row.Scan(&options.ServerID, &options.AnnounceChannel, &options.AnnounceRole, &options.Playstation, &options.Xbox, &options.Nintendo, &options.PC, &options.Awards)
	if scanErr != nil {
		return Options{}, scanErr
	}
	return options, nil
}

// check if a server is in the settings table, if not add it with default options
func checkOptions(serverID string) error {
	db, openErr := sql.Open("sqlite3", utils.DBFile)
	if openErr != nil {
		return openErr
	}
	defer db.Close()

	rows, getErr := db.Query("select server_id from settings where server_id = ?", serverID)
	if getErr != nil {
		return getErr
	}
	defer rows.Close()

	count := 0
	for rows.Next() {
		count++
	}
	if count == 0 {
		setErr := SetDefaultOptions(serverID)
		if setErr != nil {
			return setErr
		}
	}
	return nil
}
